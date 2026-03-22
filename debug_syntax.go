package main

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	bash "github.com/tree-sitter/tree-sitter-bash/bindings/go"
)

func main() {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(bash.Language()))

	tests := []string{
		"if [ -f file ]; then echo yes; fi",
		"for i in 1 2 3; do echo $i; done",
		"case $var in *) echo ok ;; esac",
	}

	for _, text := range tests {
		fmt.Printf("Text: %s\n", text)
		tree := parser.ParseCtx(context.Background(), []byte(text), nil)
		if tree == nil {
			fmt.Println("Failed to parse")
			continue
		}
		printNode(tree.RootNode(), text, 0)
		fmt.Println(strings.Repeat("-", 20))
	}
}

func printNode(node *sitter.Node, text string, depth int) {
	indent := strings.Repeat("  ", depth)
	kind := node.Kind()
	content := text[node.StartByte():node.EndByte()]
	fmt.Printf("%s[%s] (%d children): %q\n", indent, kind, node.ChildCount(), content)
	for i := uint(0); i < node.ChildCount(); i++ {
		printNode(node.Child(i), text, depth+1)
	}
}
