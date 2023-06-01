FROM registry.access.redhat.com/ubi8/nodejs-18-minimal
ENV CNB_USER_ID=1000
ENV CNB_GROUP_ID=1000
ENV CNB_STACK_ID="io.buildpacks.stacks.ubi8"
ENV CNB_STACK_DESC="ubi nodejs minimal run image base"
