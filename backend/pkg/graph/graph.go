package graph

type Graph struct {
	adj       map[int][]int
	n         int
	visited   map[int]struct{}
	cur       int
	component *Component
}

type Component struct {
	Edges []int
	Nodes []int
}

func New(nodes int) *Graph {
	return &Graph{
		adj:     make(map[int][]int),
		n:       nodes,
		visited: make(map[int]struct{}),
		cur:     1,
		component: &Component{
			Edges: []int{},
			Nodes: []int{},
		},
	}
}

func (g *Graph) AddEdge(u, v int) {
	if u == v {
		return
	}
	if _, ok := g.adj[u]; !ok {
		g.adj[u] = make([]int, 0)
	}
	if _, ok := g.adj[v]; !ok {
		g.adj[v] = make([]int, 0)
	}
	g.adj[u] = append(g.adj[u], v)
	g.adj[v] = append(g.adj[v], u)
}

func (g *Graph) dfs(v int) {
	g.visited[v] = struct{}{}
	g.component.Nodes = append(g.component.Nodes, v)
	for _, to := range g.adj[v] {
		if _, ok := g.visited[to]; !ok {
			g.component.Edges = append(g.component.Edges, v)
			g.component.Edges = append(g.component.Edges, to)
			g.dfs(to)
		}
	}
}

func (g *Graph) GetNextComponent() *Component {
	for ; g.cur <= g.n; g.cur++ {
		if _, ok := g.visited[g.cur]; !ok {
			g.component = &Component{
				Edges: []int{},
				Nodes: []int{},
			}
			g.dfs(g.cur)
			return g.component
		}
	}
	return nil
}
