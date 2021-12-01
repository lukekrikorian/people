package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"people/data"
	"strings"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "people",
		Usage: "relationship graphs for everyone else",
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "Add a user to the database. Format is ID FULL_NAME SUBTITLE. If FULL_NAME is -, interpreted from id",
				Action: func(c *cli.Context) error {
					args := c.Args()
					return AddUser(args.Get(0), args.Get(1), args.Get(2))
				},
			},
			{
				Name:  "draw",
				Usage: "Outputs the chart. Format is FILETYPE (default pdf)",
				Action: func(c *cli.Context) error {
					filetype := c.Args().First()

					if filetype == "" {
						filetype = "pdf"
					}

					data.Export(filetype)
					return nil
				},
			},
			{
				Name:  "edit",
				Usage: "Edit a user. Format is CURRENT_ID NEW_ID FULL_NAME SUBTITLE",
				Action: func(c *cli.Context) error {
					args := c.Args()
					oldid, newid := args.Get(0), args.Get(1)

					for i, person := range data.People.List {
						if person.ID == oldid {
							data.People.List = append(data.People.List[:i], data.People.List[i+1:]...)
						}
					}

					if oldid != newid {
						for i, connection := range data.People.Connections {
							for j, id := range connection.From {
								if id == oldid {
									data.People.Connections[i].From[j] = newid
								}
							}

							for j, id := range connection.To {
								if id == oldid {
									data.People.Connections[i].To[j] = newid
								}
							}
						}
					}

					return AddUser(newid, args.Get(2), args.Get(3))
				},
			},
			{
				Name:  "join",
				Usage: "Join people together. Format is COMMA_SEPARATED_IDS LABEL (\">\" or \"-\") COMMA_SEPARATED_IDS",
				Action: func(c *cli.Context) error {
					a := c.Args()
					from, label, arrowhead, to := a.Get(0), a.Get(1), a.Get(2), a.Get(3)

					var IDs []data.ID

					if to != "" {
						IDs = strings.Split(from+","+to, ",")
					} else {
						IDs = strings.Split(from, ",")
					}

					for _, id := range IDs {
						if !data.BeingUsed(id) {
							return errors.New(fmt.Sprintf("%s isn't an ID that exists.", id))
						}
					}

					data.People.Connections = append(data.People.Connections, data.Connection{
						From:      strings.Split(from, ","),
						To:        strings.Split(to, ","),
						Label:     label,
						ArrowHead: arrowhead,
					})

					data.Save()

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func MakeUser(id, name, subtitle string) (data.Person, error) {
	if data.BeingUsed(id) {
		return data.Person{}, errors.New("Already using that ID")
	}

	if name == "-" {
		name = strings.Title(strings.Join(strings.Split(id, "_"), " "))
	}

	return data.Person{
		ID:       id,
		Name:     name,
		Subtitle: subtitle,
	}, nil
}

func AddUser(id, name, subtitle string) error {
	user, err := MakeUser(id, name, subtitle)
	if err != nil {
		return err
	}

	data.People.List = append(data.People.List, user)

	data.Save()

	return nil
}

func remove(s []int, i int) []int {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
