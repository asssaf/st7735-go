: ${IMAGE_NAME:=asssaf/st7735:latest}
BASE="$(dirname $0)/.."
docker build -t $IMAGE_NAME -f $BASE/docker/Dockerfile $BASE
