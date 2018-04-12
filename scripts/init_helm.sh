# Initialize helm
if [ "$HELM_REPO_URL" != "" ] && [ "$HELM_REPO_NAME" != "" ]; then
    helm init --client-only && \
    helm repo add $HELM_REPO_NAME $HELM_REPO_URL && \
    helm repo update
fi
# Start helmi
helmi