Compile and install the binary:

```
go install github.com/drone/plugin
```

Execute a Drone plugin:

```
PLUGIN_URLS=https://webhook.site/05a1f5dc-ee5e-4c3f-86af-de39feee929a
DRONE_REPO_OWNER=octocat
DRONE_REPO_NAME=hello-world
DRONE_COMMIT_SHA=7fd1a60b01f91b314f59955a4e4d4e80d8edf11d
DRONE_COMMIT_BRANCH=master
DRONE_COMMIT_AUTHOR=octocat
DRONE_BUILD_NUMBER=1
DRONE_BUILD_STATUS=success
DRONE_BUILD_LINK=http://github.com/octocat/hello-world
DRONE_TAG=1.0.0

plugin -repo https://github.com/drone-plugins/drone-webhook.git
```

Execute a Bitrise plugin:

```
plugin -repo https://github.com/bradrydzewski/test-step.git -ref main
```
