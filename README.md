Waypoint integration with Heroku platform. Implements several different build and deployment techniques supported by Heroku so you can mix and match Waypoint plugins to support your workflows

## Install

To install the plugin, run the following command:

```bash
$ make install
```

The plugin binary is installed in `${HOME}/.config/waypoint/plugins/`.

## Use Case: Build on and Deploy Code to Heroku

Uses Heroku for builds and deployments, just orchestrated by Waypoint so you can integrate it with your other Waypoint development workflows.

```hcl
project = "example-nodejs"

app "example-nodejs" {
  labels = {
    "service" = "example-nodejs",
    "env" = "dev"
  }

  build {
    use "heroku" {
      from   = "source"
      source = "./"
      app = "example-nodejs"
    }
  }
}
```

## Use Case: Deploy Pre-Built Code to Heroku

Great for static sites or pre-compiled apps. Does not run a buildpack and quickly converts source to a deployed slug.

```hcl
project = "example-nodejs"

app "example-nodejs" {
  labels = {
    "service" = "example-nodejs",
    "env" = "dev"
  }

  build {
    use "heroku" {
      from   = "archive"
      source = "./"
      app = "example-nodejs"
    }
  }
}
```

## Use Case: Deploy Container Image to Heroku

Create a Docker container image using any Waypoint plugin ("pack" and "docker-pull" are great), then upload to Heroku Container Registry and release.

```hcl
project = "example-nodejs"

app "example-nodejs" {
  labels = {
    "service" = "example-nodejs",
    "env" = "dev"
  }

  build {
    use "pack" {}
    registry {
      use "heroku" {
        app = "example-nodejs"
      }
    }
  }

  deploy {
    use "heroku" {
        app = "example-nodejs"
    }
  }
}
```

### Build

The build stage takes application source code and converts it to and artifact, optionally pushing to a registry so it's available for the deployment platform. Heroku offers a number of ways to build code for deployment to the platform.

- Noop - let Heroku build code pushed to GitHub via its GitHub > Auto Deploy functionality. (preferred)
- From Source - Uses Heroku's programmatic build endpoint to build the local copy of the code on their servers using their slug compiler.
- From Artifact - Packages up a pre-built set of files into a "slug" on the Heroku platform
- From Container - Pushes built container to Heroku Container Registry

Heroku requires either a slug or a container to release code to an application. We can either combine the build and deploy phases and build on an indepedent app each time, or we can create one "build app" to own these slugs and containers, but it's considered independent from the apps deployed at each version.

### Deploy

Takes previously built slug or container and stages it onto Heroku.

- Noop - use latest slug on existing Heroku app
- Release slug or container image onto existing app
- Create new pipeline app from slug or container image

### Release

Activates previously staged deployment

- Noop - single-app model, or rely on Heroku Pipeline Promotion to release to a "production" app and general traffic

## Possible future features

- Integration with Heroku logs/run once Waypoint opens that up
- Support for creating a new pipeline app per deployment
