import React from "react";
import { Form, Button, Header } from "semantic-ui-react";
import backendServer from "./const";


function RemoteTunnel(props) {
    const tarPort = React.createRef()
    const handleSubmit = (target) => {
        const tar = Number(tarPort.current.value)
        if (tar > 0) {
            fetch(backendServer.concat("/tunnel/remote"), {
                method: "post",
                body: {
                    tar: tar,
                }
            }).
                then(response => response.json()).
                then(response => {
                    if (response.err.length > 0) {
                        alert("Error Occured while creating remote tunnel: ", response.err)
                    } else {
                        alert("Successfully Created Remote Tunnel!!")
                        tarPort.current.value = ""
                    }
                })

        } else {
            alert("Invalid Port(s)")
        }
    }
    return (
        <div>
            <Header>Remote Tunnel</Header>

            <Form>
                <Form.Field>
                    <label>Port to Tunnel</label>
                    <input placeholder="port" ref={tarPort} />
                </Form.Field>
                <Button onClick={handleSubmit}>Create Tunnel</Button>
            </Form>
        </div>
    )
}

export default RemoteTunnel;