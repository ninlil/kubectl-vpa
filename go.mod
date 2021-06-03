module github.com/ninlil/kubectl-vpa

go 1.15

require (
	cloud.google.com/go v0.54.0 // indirect
	github.com/Azure/go-autorest/autorest v0.11.1 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.5 // indirect
	github.com/alexflint/go-arg v1.3.0
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/imdario/mergo v0.3.5 // indirect
	github.com/ninlil/ansi v1.1.0 // indirect
	github.com/ninlil/columns v1.0.0 // indirect
	github.com/ninlil/vpa-compare/vpa_v1 v0.0.0
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	k8s.io/api v0.20.5
	k8s.io/apimachinery v0.20.5
	k8s.io/client-go v0.20.5
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
)

replace github.com/ninlil/vpa-compare/vpa_v1 => ./vpa_v1

replace github.com/ninlil/columns => ../columns
