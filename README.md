Compile and install the binary:

```
go install github.com/drone/plugin
```

Execute a Harness plugin:

```
export PLUGIN_URLS=https://webhook.site/05a1f5dc-ee5e-4c3f-86af-de39feee929a
export DRONE_REPO_OWNER=octocat
export DRONE_REPO_NAME=hello-world
export DRONE_COMMIT_SHA=7fd1a60b01f91b314f59955a4e4d4e80d8edf11d
export DRONE_COMMIT_BRANCH=master
export DRONE_COMMIT_AUTHOR=octocat
export DRONE_BUILD_NUMBER=1
export DRONE_BUILD_STATUS=success
export DRONE_BUILD_LINK=http://github.com/octocat/hello-world
export DRONE_TAG=1.0.0

plugin -repo https://github.com/drone-plugins/drone-webhook.git
```

Execute a Harness plugin by alias:

```
plugin -name webhook
```

Execute a Bitrise plugin:

```
plugin -repo https://github.com/bradrydzewski/test-step.git -ref main
```
