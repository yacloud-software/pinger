package dot

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"strings"

	"github.com/goccy/go-graphviz"
)

type node struct {
	Name    string
	Targets []*nodetarget
}
type nodetarget struct {
	Name string
	ok   bool
}

func GenerateDot(nodes []*node) (string, error) {
	res := `digraph callgraph {
 graph [dpi=300,layout="dot"];
 node[fontname="FreeSans",fontsize=20];
 subgraph info {
   foo[label="Generated with pinger-client ",shape=record,color=yellow,style=filled];
 }
 rankdir=LR;
 subgraph modules {
`
	for _, n := range nodes {
		for _, t := range n.Targets {
			col := "red"
			if t.ok {
				col = `"#55ff55"`
			}
			name := strings.ReplaceAll(t.Name, ".", "_")
			name = strings.ReplaceAll(name, "-", "_")
			line := "   " + n.Name + " -> \"" + name + "\" [color=" + col + "];\n"
			res = res + line
		}
	}
	res = res + `
  }
}
`
	return res, nil
}

func GenerateImage(dot string) (image.Image, error) {
	ctx := context.Background()
	g, err := graphviz.New(ctx)
	if err != nil {
		panic(err)
	}

	graph, err := graphviz.ParseBytes([]byte(dot))
	if err != nil {
		return nil, err
	}

	var gbuf bytes.Buffer

	err = g.Render(ctx, graph, graphviz.PNG, &gbuf)
	if err != nil {
		return nil, err
	}

	// 2. get as image.Image instance
	image, err := g.RenderImage(ctx, graph)
	if err != nil {
		return nil, err
	}
	return image, nil
}
func GeneratePNG(dot string) ([]byte, error) {
	img, err := GenerateImage(dot)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	err = png.Encode(buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
