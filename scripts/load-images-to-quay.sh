rm ../build/*.oci
rm ../build-nodejs-16/*.oci
rm ../build-nodejs-18/*.oci
bash ./create.sh
bash ./create-nodejs-16.sh
bash ./create-nodejs-18.sh

skopeo copy --dest-tls-verify=false oci-archive:../build/run.oci docker-daemon:quay.io/midawson/ubi8-paketo-run:latest
skopeo copy --dest-tls-verify=false oci-archive:../build/run.oci docker-daemon:quay.io/midawson/ubi8-paketo-build:latest
skopeo copy --dest-tls-verify=false oci-archive:../build-nodejs-16/run.oci docker-daemon:quay.io/midawson/ubi8-paketo-run-nodejs-16:latest
skopeo copy --dest-tls-verify=false oci-archive:../build-nodejs-18/run.oci docker-daemon:quay.io/midawson/ubi8-paketo-run-nodejs-18:latest

docker push quay.io/midawson/ubi8-paketo-build
docker push quay.io/midawson/ubi8-paketo-run
docker push quay.io/midawson/ubi8-paketo-run-nodejs-16
docker push quay.io/midawson/ubi8-paketo-run-nodejs-18

