package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type ID = string

var (
	People Ppl
	file   *os.File
	err    error
)

const template = `
digraph people {
	layout=dot
	rankdir=TB
	overlap=false
	epsilon=1
	splines=true
	nodesep=0.3
	ranksep=0.3
	node[shape=box];
	edge[arrowsize=0.5];
	%s

	node[shape=plaintext];
	%s

	%s
}
`

type Ppl struct {
	List        []Person     `json:"people"`
	Connections []Connection `json:"connections"`
	labels      []string
	counter     int
}

type Person struct {
	ID       ID
	Name     string `json:"name"`
	Subtitle string `json:"subtitle"`
}

type Connection struct {
	From      []ID   `json:"from"`
	To        []ID   `json:"to"`
	Label     string `json:"label"`
	ArrowHead string `json:"arrow_head"`
}

func Joined(arr []ID) string {
	return strings.Join(arr, ", ")
}

func (c *Connection) GenerateEdge() (s string) {
	People.labels = append(People.labels, fmt.Sprintf("%d [label=\"%s\"]\n\t", People.counter, c.Label))

	s += fmt.Sprintf("{%s} -> %d", Joined(c.From), People.counter)

	if len(c.To) > 0 {
		s += fmt.Sprintf(" -> {%s}", Joined(c.To))
	}

	if c.ArrowHead == "-" {
		s += "[dir=both,arrowsize=0]"
	}

	People.counter += 1

	return s + "\n\t"
}

func (p *Person) GenerateNode() (s string) {
	var extra string

	if p.Subtitle != "" {
		extra = fmt.Sprintf("<br/><font point-size=\"12\">%s</font>", p.Subtitle)
	}

	return fmt.Sprintf("%s [label=<%s%s>]\n\t", p.ID, p.Name, extra)
}

func (p *Ppl) GenerateGraph() string {
	var ps, cs string

	for _, person := range p.List {
		ps += person.GenerateNode()
	}

	for _, connection := range p.Connections {
		cs += connection.GenerateEdge()
	}

	return fmt.Sprintf(template, ps, strings.Join(People.labels, ""), cs)
}

func BeingUsed(ID string) bool {
	for _, person := range People.List {
		if person.ID == ID {
			return true
		}
	}
	return false
}

func init() {
	path, _ := os.Getwd()
	file, err = os.OpenFile(path+"/people.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Couldn't find people.json file")
		os.Exit(1)
	}

	parser := json.NewDecoder(file)
	err = parser.Decode(&People)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Save() {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "\t")
	err = enc.Encode(People)

	if err != nil {
		fmt.Println(err)
	}

	file.Truncate(0)
	file.Seek(0, 0)
	file.Write(buf.Bytes())
}

func Export(filetype string) {
	os.WriteFile("people.dot", []byte(People.GenerateGraph()), 0644)
	cmd := exec.Command("dot", "-T"+filetype, "people.dot")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.WriteFile("people."+filetype, stdout, 0644)
}
