"""Defines the dependencies' versions.

This allow us to keep the book generic and update cheaply the versions.
"""

# Go
# The last version can be found here: https://go.dev/dl/
# Note: You do not need to download the last version locally,
#       Just update this variable. Bazel will download it for you.
GO_VERSION = "1.23.0"

# Rules_go
# The last version and SHA256 can be found here: https://github.com/bazelbuild/rules_go/releases
RULES_GO_VERSION = "v0.49.0"
RULES_GO_SHA256 = "d93ef02f1e72c82d8bb3d5169519b36167b33cf68c252525e3b9d3d5dd143de7"

# Gazelle
# The last version and SHA256 can be found here: https://github.com/bazelbuild/bazel-gazelle/blob/master/README.rst
GAZELLE_VERSION = "v0.35.0"
GAZELLE_SHA256 = "32938bda16e6700063035479063d9d24c60eda8d79fd4739563f50d331cb3209"

# Protobuf
# The last version can be found here: https://github.com/protocolbuffers/protobuf/releases
# Note: You do not need to download the last version locally,
#       Just update this variable. Bazel will download it for you.
PROTO_VERSION = "v27.0"

# protoc-gen-validate
# The last version can be found here: https://github.com/bufbuild/protoc-gen-validate/releases
PROTOC_GEN_VALIDATE_VERSION = "v1.1.0"