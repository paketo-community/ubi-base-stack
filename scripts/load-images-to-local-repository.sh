rm ../build/*.oci
rm ../build-nodejs-16/*.oci
rm ../build-nodejs-18/*.oci
bash ./create.sh
bash ./create-nodejs-16.sh
bash ./create-nodejs-18.sh
docker images |grep ubi8-paketo |  awk  '{print "docker rmi --force " $3}' |bash
skopeo copy --dest-tls-verify=false oci-archive:../build/run.oci docker-daemon:localhost:5000/ubi8-paketo-run:latest
skopeo copy --dest-tls-verify=false oci-archive:../build/run.oci docker-daemon:localhost:5000/ubi8-paketo-build:latest
skopeo copy --dest-tls-verify=false oci-archive:../build-nodejs-16/run.oci docker-daemon:localhost:5000/ubi8-paketo-run-nodejs-16:latest
skopeo copy --dest-tls-verify=false oci-archive:../build-nodejs-18/run.oci docker-daemon:localhost:5000/ubi8-paketo-run-nodejs-18:latest

