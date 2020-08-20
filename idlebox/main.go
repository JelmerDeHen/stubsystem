package main

import (
	"fmt"
	"os"
	"github.com/JelmerDeHen/stubsystem/idlebox/cmds"
	"strings"
)
/*
Execution lays a bit more complicated than this

The functions should probably look like:
cmds.Cmd(fds []*os.File, environ []string, pwd string, argv []string)

Then the fds are connected to pipes/streams/files:
fd[0] os.Stdin
fd[1] /dev/stdout
fd[2] /dev/stderr
fd[x] ...

When user does:
echo "test" >&2

POSIX guidelines
https://www.gnu.org/software/libc/manual/html_node/Argument-Syntax.html

*/
func main() {
	cmd := strings.TrimPrefix(os.Args[0], "./")
	cmdSplit := strings.Split(cmd, "/")
	if len(cmdSplit) > 0 {
		cmd = cmdSplit[len(cmdSplit)-1]
	}
	argv := os.Args[1:]

	cmdlist := []string{"cat", "insmod", "ls", "cd"}
	switch cmd {
	case "cat":
		retcode := cmds.Cat(argv)
		os.Exit(retcode)
	case "insmod":
		retcode := cmds.Insmod(argv)
		os.Exit(retcode)
	case "ls":
		retcode := cmds.Ls(argv)
		os.Exit(retcode)
	case "cd":
		// cd is shell builtin
		// This pwd will not affect the
		if len(argv) == 1 {
			os.Chdir(argv[0])
		}
	default:
		if len(argv) != 0 {
			if argv[0] == "--list" {
				fmt.Printf("%s\n", strings.Join(cmdlist, "\n"))
				return
			}
		}
		fmt.Println("Usage: exec -a <cat,insmod> ./idlebox")
	}
	return
}
