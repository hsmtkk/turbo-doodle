up-pod:
	podman play kube pod.yml

down-pod:
	podman pod rm -f turbo-doodle
