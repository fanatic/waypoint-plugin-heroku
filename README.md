# waypoint-plugin-heroku

Waypoint integration with Heroku platform

**WORK IN PROGRESS: this plugin is not yet functional**

## Build

The build stage takes application source code and converts it to and artifact, optionally pushing to a registry so it's available for the deployment platform. Heroku offers a number of ways to build code for deployment to the platform.

- Noop - let Heroku build code pushed to GitHub via its GitHub > Auto Deploy functionality. (preferred)
- From Source - Uses Heroku's programmatic build endpoint to build the local copy of the code on their servers using their slug compiler.
- From Artifact - Packages up a pre-built set of files into a "slug" on the Heroku platform
- From Container - Pushes built container to Heroku Container Registry

Heroku requires either a slug or a container to release code to an application. We can either combine the build and deploy phases and build on an indepedent app each time, or we can create one "build app" to own these slugs and containers, but it's considered independent from the apps deployed at each version.

## Deploy

Takes previously built slug or container and stages it onto Heroku.

- Noop - use latest slug on existing Heroku app
- Create new pipeline app from the given artifact

## Release

Activates previously staged deployment

- Noop - single-app model, or rely on Heroku Pipeline Promotion to release to a "production" app and general traffi
- Custom Domain - Moves the public DNS domain to a staged app
