: ${IMAGE_NAME:=asssaf/st7735:latest}
docker run --rm -it --privileged --device /dev/spidev0.1 "$IMAGE_NAME" $*
