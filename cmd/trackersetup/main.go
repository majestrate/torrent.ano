package main

import (
	"fmt"
	"os"
	"os/user"
)

func printErr(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err.Error())
}

func main() {
	user, err := user.Current()
	if err != nil {
		printErr("failed to get current user", err)
		return
	}
	name := user.Username
	fmt.Fprintf(os.StdOut, "create role %s with login;\ncreate database %s with owner %s;\n", name, name, name)
}
