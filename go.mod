module github.com/ninlil/kubectl-vpa

go 1.15

require (
	github.com/alexflint/go-arg v1.4.2
	github.com/go-test/deep v1.0.7 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/mickep76/encoding v0.0.0-20191112132937-4a810d6b3199
	github.com/ninlil/ansi v1.1.0
	github.com/ninlil/columns v1.0.0
	github.com/ninlil/kubectl-vpa/vpa_v1 v0.0.0
	golang.org/x/oauth2 v0.0.0-20210514164344-f6687ab2804c // indirect
	golang.org/x/sys v0.0.0-20210603125802-9665404d3644 // indirect
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b // indirect
)

replace github.com/ninlil/kubectl-vpa/vpa_v1 => ./vpa_v1

replace github.com/ninlil/columns => ../columns
