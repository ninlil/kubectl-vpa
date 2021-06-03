module github.com/ninlil/kubectl-vpa

go 1.15

require (
	github.com/alexflint/go-arg v1.3.0
	github.com/mickep76/encoding v0.0.0-20191112132937-4a810d6b3199
	github.com/ninlil/ansi v1.1.0
	github.com/ninlil/columns v1.0.0
	github.com/ninlil/kubectl-vpa/vpa_v1 v0.0.0
	k8s.io/api v0.20.5
	k8s.io/apimachinery v0.20.5
	k8s.io/client-go v0.20.5
)

replace github.com/ninlil/kubectl-vpa/vpa_v1 => ./vpa_v1

replace github.com/ninlil/columns => ../columns
