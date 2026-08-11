# Copilot Instructions

These high-level instructions define how GitHub Copilot should assist with this Go project and should not override the more specific instructions found under `.github/instructions`. The goal is to ensure consistent, high-quality code generation aligned with Go idioms, the chosen architecture, and our team's best practices.

## Overview

This project is a Terraform provider for Kubernetes that's based on a native Server Side Apply (SSA) approach making use of schemas to allow planning against either a non-existent cluster or a cluster without the resource present yet. This provider is written in Go using the modern [`github.com/hashicorp/terraform-plugin-framework`](https://github.com/hashicorp/terraform-plugin-framework) ([docs](https://developer.hashicorp.com/terraform/plugin/framework)) and the [`k8s.io/client-go`](https://github.com/kubernetes/client-go) library.
