#!/usr/bin/env bats

load helpers

host="localregistry:5440"

function docker_noschema2() {
	docker exec dockerdaemon-noschema2 docker $@
}

@test "Test schema2 fallback to schema1" {
	version_check docker "$GOLEM_DIND_VERSION" "1.10.0"
	version_check registry "$GOLEM_DISTRIBUTION_VERSION" "2.3.0"

	imagename="$host/testschema2toschema1:latest"
	tempImage $imagename
	run docker_t push $imagename
	echo $output
	[ "$status" -eq 0 ]

	docker_t rmi $imagename

	run docker_noschema2 pull $imagename
	echo $output
	[ "$status" -eq 0 ]

}

@test "Test schema2 to schema1 pull by digest failure" {
	version_check docker "$GOLEM_DIND_VERSION" "1.10.0"
	version_check registry "$GOLEM_DISTRIBUTION_VERSION" "2.3.0"

	imagename="$host/testschema2toschema1bydigest"
	image="$imagename:latest"
	tempImage $image
	run docker_t push $image
	echo "$output"
	[ "$status" -eq 0 ]
	has_digest "$output"

	# Remove image to ensure layer is pulled and digest verified
	docker_t rmi -f $image

	run docker_noschema2 pull "$imagename@$digest"
	echo "$output"
	[ "$status" -ne 0 ]
}
