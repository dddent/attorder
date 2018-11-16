package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"strings"

	"github.com/dddent/attorder/lexxer"
	"github.com/dddent/attorder/parser"
)

func main() {
	var write bool
	flag.BoolVar(&write, "write", false, "weather to write back the result to the file")
	flag.BoolVar(&write, "w", false, "write shorthand")
	var stdio bool
	flag.BoolVar(&stdio, "stdio", false, "read from stdin, write to stdout")
	var orderArg string
	flag.StringVar(&orderArg, "order", "", "Comma seperated list of regular expressions in the order the attributes should appear")
	flag.Parse()
	order := strings.Split(orderArg, ",")
	log.Println(order)
	if stdio {
		s := parseFile(os.Stdin, order)
		writeFile(os.Stdout, s)
		return
	}
	for _, arg := range flag.Args() {
		f, err := os.Open(arg)
		if err != nil {
			panic(err)
		}
		s := parseFile(f, order)
		f.Close()
		f = os.Stdout
		if write {
			f, err = os.Create(arg)
			if err != nil {
				panic(err)
			}
		}
		writeFile(f, s)
		f.Close()
	}
}

func parseFile(f io.Reader, order []string) string {
	l := lexxer.New(f)
	p, err := parser.New(l, parser.Settings{
		AttrOrder: order,
	})
	if err != nil {
		panic(err)
	}
	a, err := p.ParseAST()
	if err != nil {
		panic(err)
	}
	return a.HTMLString()
}

func writeFile(f io.Writer, s string) {
	w := bufio.NewWriter(f)
	w.WriteString(s)
	w.Flush()
	return
}
