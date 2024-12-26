resource "aws_security_group" "elb_sg" {
  name        = "elb-sg"
  description = "Security froup for ELB instances"
  vpc_id      = aws_default_vpc.default.id

  ingress {
    description = "Allow HTTP traffic from anywhere"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "Allow all outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_lb" "elb" {
  name               = "elb"
  load_balancer_type = "application"
  security_groups    = [aws_security_group.elb_sg.id]
  subnets            = values(aws_default_subnet.default)[*].id
}

resource "aws_lb_target_group" "elb_tg" {
  name     = "elb-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_default_vpc.default.id
  health_check {
    enabled = true
    path    = "/ping"
  }
}

resource "aws_lb_target_group_attachment" "elb_tg_attachment" {
  count            = var.instance_count
  target_group_arn = aws_lb_target_group.elb_tg.arn
  target_id        = aws_instance.instance[count.index].id
  port             = 80
}

resource "aws_lb_listener" "elb_listener" {
  load_balancer_arn = aws_lb.elb.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_lb_target_group.elb_tg.arn
    type             = "forward"
  }
}
