DOCKER_ARCHS ?= amd64

include Makefile.common
L1 := data
L2 := eventbus
GIT_REPO := craftDemo
CODECOV_TOKEN := b85a2363-8ef4-4794-841d-babb996aa0cc

# use semantic versioning
# add more logic to automate: either replace or update version
THIS_PACKAGE_VERSION ?= 0.0.1
DOCKER_IMAGE_NAME       ?= $(GIT_REPO)

