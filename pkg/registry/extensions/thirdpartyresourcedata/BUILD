package(default_visibility = ["//visibility:public"])

licenses(["notice"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_binary",
    "go_library",
    "go_test",
    "cgo_library",
)

go_library(
    name = "go_default_library",
    srcs = [
        "codec.go",
        "doc.go",
        "registry.go",
        "strategy.go",
        "util.go",
    ],
    tags = ["automanaged"],
    deps = [
        "//pkg/api:go_default_library",
        "//pkg/api/meta:go_default_library",
        "//pkg/api/rest:go_default_library",
        "//pkg/api/unversioned:go_default_library",
        "//pkg/api/util:go_default_library",
        "//pkg/api/v1:go_default_library",
        "//pkg/apimachinery/registered:go_default_library",
        "//pkg/apis/extensions:go_default_library",
        "//pkg/apis/extensions/v1beta1:go_default_library",
        "//pkg/apis/extensions/validation:go_default_library",
        "//pkg/fields:go_default_library",
        "//pkg/labels:go_default_library",
        "//pkg/runtime:go_default_library",
        "//pkg/storage:go_default_library",
        "//pkg/util/validation/field:go_default_library",
        "//pkg/util/yaml:go_default_library",
        "//pkg/watch:go_default_library",
        "//pkg/watch/versioned:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "codec_test.go",
        "strategy_test.go",
        "util_test.go",
    ],
    library = "go_default_library",
    tags = ["automanaged"],
    deps = [
        "//pkg/api:go_default_library",
        "//pkg/api/testapi:go_default_library",
        "//pkg/api/testing:go_default_library",
        "//pkg/api/unversioned:go_default_library",
        "//pkg/api/v1:go_default_library",
        "//pkg/apimachinery/registered:go_default_library",
        "//pkg/apis/extensions:go_default_library",
        "//pkg/runtime:go_default_library",
        "//pkg/watch/versioned:go_default_library",
    ],
)
