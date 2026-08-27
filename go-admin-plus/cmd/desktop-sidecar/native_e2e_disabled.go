//go:build !desktop_native_e2e

package main

import "net/http"

func (*sidecarRuntime) registerNativeE2EControl(*http.ServeMux) {}
