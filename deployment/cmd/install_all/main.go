package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

type Instance struct {
	IP        string `json:"ip"`
	IsGateway bool   `json:"is_gateway"`
	IsManager bool   `json:"is_manager"`
}

type TerraformOutput struct {
	Instances struct {
		Value []Instance `json:"value"`
	} `json:"instances"`
	SSHKey struct {
		Value string `json:"value"`
	} `json:"ssh_key"`
}

func runCommands(client *ssh.Client, commands ...string) error {
	for _, command := range commands {
		session, err := client.NewSession()
		if err != nil {
			return fmt.Errorf("create ssh session: %w", err)
		}
		result, err := session.CombinedOutput(command)
		session.Close()
		if err != nil {
			return fmt.Errorf("ssh session: %s: %s; %w", command, result, err)
		}
	}
	return nil
}

func extractJoinCommand(command string) string {
	command = strings.TrimRight(command, "\n")
	startIndex := strings.Index(command, "docker swarm join")
	return command[startIndex:]
}

func main() {
	var t TerraformOutput
	json.NewDecoder(os.Stdin).Decode(&t)
	instances := t.Instances.Value
	ssh_key := t.SSHKey.Value

	for _, instance := range instances {
		if instance.IsGateway && !instance.IsManager {
			log.Fatalf("instance %v: gateway instance is supposed to be a Swarm manager", instance)
		}
	}

	signer, err := ssh.ParsePrivateKey([]byte(ssh_key))
	if err != nil {
		log.Fatal(err)
	}

	clients := make([]*ssh.Client, len(instances))
	defer func() {
		for _, client := range clients {
			if client != nil {
				client.Close()
			}
		}
	}()
	g := &errgroup.Group{}
	for i, instance := range instances {
		config := &ssh.ClientConfig{
			User:            "ec2-user",
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		g.Go(func(i int, instance Instance) func() error {
			return func() error {
				var err error
				clients[i], err = ssh.Dial("tcp", fmt.Sprintf("%s:22", instance.IP), config)
				if err != nil {
					return err
				}
				err = runCommands(
					clients[i],
					"sudo yum update -y",
					"sudo yum install git -y",
					"rm -rf smapp",
					"git clone https://github.com/avoronoi/smapp",
					"./smapp/deployment/scripts/install.sh",
				)
				if err != nil {
					return err
				}
				return nil
			}
		}(i, instance))
	}
	if err = g.Wait(); err != nil {
		log.Fatal(err)
	}

	initSwarmNode := -1
	for i, instance := range instances {
		if instance.IsManager {
			initSwarmNode = i
			break
		}
	}
	if initSwarmNode == -1 {
		log.Fatal("expected at least one manager instance")
	}

	fmt.Println("Init swarm on", instances[initSwarmNode].IP)
	err = runCommands(clients[initSwarmNode], "./smapp/deployment/scripts/init_swarm.sh")
	if err != nil {
		log.Fatal(err)
	}

	joinCommand := make(map[string]string)
	for _, nodeType := range []string{"worker", "manager"} {
		command := fmt.Sprintf("docker swarm join-token %s", nodeType)
		session, err := clients[initSwarmNode].NewSession()
		if err != nil {
			log.Fatalf("create ssh session: %s", err)
		}
		initSwarmOutput, err := session.Output(command)
		session.Close()
		if err != nil {
			log.Fatalf("ssh session: %s: %s", command, err)
		}
		joinCommand[nodeType] = extractJoinCommand(string(initSwarmOutput))
	}

	for i, instance := range instances {
		if i == initSwarmNode {
			continue
		}
		g.Go(func(i int, instance Instance) func() error {
			return func() error {
				var command string
				if instance.IsManager {
					command = joinCommand["manager"]
				} else {
					command = joinCommand["worker"]
				}
				fmt.Println("apply", command, "on instance", instance.IP)
				return runCommands(clients[i], command)
			}
		}(i, instance))
	}
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}

	for i, instance := range instances {
		if !instance.IsGateway {
			continue
		}
		g.Go(func(i int, instance Instance) func() error {
			return func() error {
				fmt.Println("set gateway on instance", instance.IP)
				return runCommands(clients[i], "./smapp/deployment/scripts/set_gateway.sh")
			}
		}(i, instance))
	}
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
