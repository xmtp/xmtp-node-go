import sys
import getopt
import urllib.request
import json

TERRAFORM_BASE_URL = "https://app.terraform.io/api/v2"
NODE_IMAGE_PREFIX = "xmtp/node-go@sha256:"


class Terraform:
    def __init__(self, api_token, organization, workspace_name):
        self.api_token = api_token
        self.organization = organization
        self.workspace_name = workspace_name
        self.workspace_id = self._find_workspace_id(workspace_name)

    def _terraform_request(self, path, data=None, method="GET"):
        req = urllib.request.Request(
            f"{TERRAFORM_BASE_URL}{path}",
            headers={"Content-Type": "application/vnd.api+json", "Authorization": f"Bearer {self.api_token}"},
            method=method,
        )

        if data != None:
            req.data = json.dumps(data).encode()

        print(f"Doing request at {path} with data {data}")

        with urllib.request.urlopen(req) as response:
            body = response.read().decode(response.headers.get_content_charset("utf-8"))
            return json.loads(body)["data"]

    def _find_workspace_id(self, workspace_name):
        workspaces = self._terraform_request(f"/organizations/{self.organization}/workspaces")
        for workspace in workspaces:
            if workspace["attributes"]["name"] == workspace_name:
                return workspace["id"]

        raise Exception(f"Workspace {self.workspace} not found")

    def _find_variable(self, key):
        variables = self._terraform_request(f"/workspaces/{self.workspace_id}/vars")
        for variable in variables:
            if variable["attributes"]["key"] == key:
                return variable
        raise Exception("Could not find variable")

    def set_workspace_variable(self, key, val):
        variable = self._find_variable(key)
        print(f"Setting {key} to {val} in workspace {self.workspace_name}")
        body = {"data": {"type": "vars", "id": variable["id"], "attributes": {"key": key, "value": val}}}
        return self._terraform_request(f"/workspaces/{self.workspace_id}/vars/{variable['id']}", body, method="PATCH")

    def start_run(self, git_commit):
        body = {
            "data": {
                "type": "runs",
                "attributes": {"message": f"Triggered from xmtp-node-go commit {git_commit}"},
                "relationships": {"workspace": {"data": {"type": "workspaces", "id": self.workspace_id}}},
            }
        }
        run = self._terraform_request("/runs", body, method="POST")
        return run


if __name__ == "__main__":
    opts, _ = getopt.getopt(
        sys.argv[1:], "", ["tf-token=", "workspace=", "organization=", "xmtp-node-image=", "git-commit="]
    )
    opts = {k.replace("--", ""): v for k, v in dict(opts).items()}
    tf = Terraform(opts["tf-token"], opts["organization"], opts["workspace"])

    node_image = opts["xmtp-node-image"].strip()
    if not node_image.startswith(NODE_IMAGE_PREFIX):
        raise Exception(f"Invalid node image {node_image}")

    tf.set_workspace_variable("xmtp_node_image", node_image)
    tf.start_run(opts["git-commit"])
