package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ojrac/libmigrate"
)

func run(m libmigrate.Migrator, args []string) {
	err := doRun(m, args)

	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}

func doRun(m libmigrate.Migrator, args []string) (err error) {
	ctx := context.Background()

	if len(args) == 0 {
		flag.CommandLine.Usage()
		return nil
	}

	command := args[0]
	if version, err := strconv.Atoi(command); err == nil {
		return m.MigrateTo(ctx, version)
	}

	switch command {
	case "latest":
		return m.MigrateLatest(ctx)
	case "create":
		if len(args) < 2 {
			fmt.Printf("create: Missing migration name\n\n")
			flag.CommandLine.Usage()
			os.Exit(1)
			return
		}
		name := args[1]
		return m.Create(ctx, name)
	case "version":
		version, err := m.GetVersion(ctx)
		if err == nil {
			fmt.Printf("%d\n", version)
		}
		return err
	case "tool-version":
		fmt.Printf("%s\n", ToolVersion)
		return nil
	case "pending":
		hasPending, err := m.HasPending(ctx)
		if err == nil {
			if hasPending {
				fmt.Printf("true\n")
			} else {
				fmt.Printf("false\n")
			}
		}
		return err
	}

	flag.CommandLine.Usage()
	os.Exit(1)
	return
}

func parseEnv(vars map[string]*string) {
	env := os.Environ()

	for _, val := range env {
		parts := strings.SplitN(val, "=", 2)
		if p, ok := vars[parts[0]]; ok {
			*p = parts[1]
		}
	}
}
