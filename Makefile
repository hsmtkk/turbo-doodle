up-pod:
	podman play kube pod.yml

down-pod:
	podman pod rm -f minio_pod
