#!/usr/bin/env bats

# Registry host name, should be set to non-localhost address and match
# DNS name in nginx/ssl certificates and what is installed in /etc/docker/cert.d

load helpers

hostname=${TEST_REGISTRY:-"localregistry"}

repo=${TEST_REPO:-"hello-world"}
tag=${TEST_TAG:-"latest"}
image="${repo}:${tag}"

# Login information, should match values in nginx/test.passwd
user=${TEST_USER:-"testuser"}
password=${TEST_PASSWORD:-"passpassword"}
email="distribution@docker.com"

function setup() {
	if [ "$TEST_SKIP_PULL" == "" ]; then
		docker pull $image
	fi
}

@test "Test valid certificates" {
	docker tag -f $image $hostname:5440/$image
	run docker push $hostname:5440/$image
	[ "$status" -eq 0 ]
	has_digest "$output"
}

@test "Test basic auth" {
	basic_auth_version_check
	login $hostname:5441
	docker tag -f $image $hostname:5441/$image
	run docker push $hostname:5441/$image
	[ "$status" -eq 0 ]
	has_digest "$output"
}

@test "Test TLS client auth" {
	docker tag -f $image $hostname:5442/$image
	run docker push $hostname:5442/$image
	[ "$status" -eq 0 ]
	has_digest "$output"
}

@test "Test TLS client with invalid certificate authority fails" {
	docker tag -f $image $hostname:5443/$image
	run docker push $hostname:5443/$image
	[ "$status" -ne 0 ]
}

@test "Test basic auth with TLS client auth" {
	basic_auth_version_check
	login $hostname:5444
	docker tag -f $image $hostname:5444/$image
	run docker push $hostname:5444/$image
	[ "$status" -eq 0 ]
	has_digest "$output"
}

@test "Test unknown certificate authority fails" {
	docker tag -f $image $hostname:5445/$image
	run docker push $hostname:5445/$image
	[ "$status" -ne 0 ]
}

@test "Test basic auth with unknown certificate authority fails" {
	run login $hostname:5446
	[ "$status" -ne 0 ]
	docker tag -f $image $hostname:5446/$image
	run docker push $hostname:5446/$image
	[ "$status" -ne 0 ]
}

@test "Test TLS client auth to server with unknown certificate authority fails" {
	docker tag -f $image $hostname:5447/$image
	run docker push $hostname:5447/$image
	[ "$status" -ne 0 ]
}

@test "Test failure to connect to server fails to fallback to SSLv3" {
	docker tag -f $image $hostname:5448/$image
	run docker push $hostname:5448/$image
	[ "$status" -ne 0 ]
}

