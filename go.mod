module github.com/fanatic/waypoint-plugin-heroku

go 1.15

require (
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d
	github.com/docker/cli v0.0.0-20200312141509-ef2f64abbd37
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.4.2-0.20200221181110-62bd5a33f707
	github.com/golang/protobuf v1.4.2
	github.com/hashicorp/go-hclog v0.14.1
	github.com/hashicorp/waypoint v0.1.3
	github.com/hashicorp/waypoint-plugin-examples/template v0.0.0-20201015155043-8dca3e6761cf
	github.com/hashicorp/waypoint-plugin-sdk v0.0.0-20201016002013-59421183d54f
	github.com/heroku/heroku-go/v5 v5.2.0
	github.com/jdxcode/netrc v0.0.0-20190329161231-b36f1c51d91d
	github.com/paketo-buildpacks/procfile v1.4.0
	github.com/rs/zerolog v1.20.0
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.25.0
)

replace (
	// v0.3.11 panics for some reason on our tests
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.9

	// https://github.com/ory/dockertest/issues/208
	golang.org/x/sys => golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6
)
