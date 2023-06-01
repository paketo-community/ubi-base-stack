FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
ENV CNB_USER_ID=1002
ENV CNB_GROUP_ID=1000
ENV CNB_STACK_ID="io.buildpacks.stacks.ubi8"
ENV CNB_STACK_DESC="ubi nodejs minimal run image base"
