package gircd

import (
	"bufio"
	"fmt"
	"strings"
)

type ircCmd struct {
	src     string
	cmd     string
	params  []string
	trailer string
}

func auth(s *bufio.Scanner) (*user, error) {
	var username, realname, nick, password string
	for s.Scan() {
		cmd, err := ircParse(s.Text())
		if err != nil {
			return nil, err
		}
		switch cmd.cmd {
		case "USER":
			if len(cmd.params) > 1 {
				username = cmd.params[0]
				realname = cmd.trailer
			}
		case "NICK":
			if len(cmd.params) == 1 {
				nick = cmd.params[0]
			}
		case "PASS":
			if len(cmd.params) == 1 {
				password = cmd.params[0]
			}
		default:
			return nil, fmt.Errorf("bad auth")
		}
		if len(username) > 0 && len(nick) > 0 {
			return &user{
				username: username,
				realname: realname,
				password: password,
				nick:     nick,
			}, nil
		}
	}
	if s.Err() != nil {
		return nil, s.Err()
	}
	return nil, fmt.Errorf("weird auth")
}

func ircParse(l string) (*ircCmd, error) {
	if len(l) <= 1 {
		return &ircCmd{cmd: "NOOP"}, nil
	}
	haveSrc := false
	if l[0] == ':' {
		l = l[1:]
		haveSrc = true
	}
	a := strings.SplitN(l, ":", 2)
	trailer := ""
	if len(a) == 2 {
		trailer = strings.TrimSpace(a[1])
		l = a[0]
	}
	if len(l) == 0 {
		return &ircCmd{cmd: "NOOP"}, nil
	}
	a = strings.Split(l, " ")
	src := ""
	if haveSrc {
		src = strings.TrimSpace(a[0])
		a = a[1:]
	}
	cmd := strings.ToUpper(strings.TrimSpace(a[0]))
	if len(a) == 0 {
		return &ircCmd{cmd: cmd}, nil
	}
	params := make([]string, 0, 16)
	for _, v := range a[1:] {
		params = append(params, strings.TrimSpace(v))
	}
	return &ircCmd{src: src, cmd: cmd, params: params, trailer: trailer}, nil
}
