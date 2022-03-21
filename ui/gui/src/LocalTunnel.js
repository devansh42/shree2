import React from "react";
import { Form, Button, Header } from "semantic-ui-react";
import backendServer from "./const";


function LocalTunnel(props) {
    const tarPort = React.createRef()
    const srcPort = React.createRef()
    const handleSubmit = (target) => {
        const src = Number(srcPort.current.value)
        const tar = Number(tarPort.current.value)
        if (src > 0 && tar > 0 && src != tar) {
            fetch(backendServer.concat("/tunnel/local"), {
                method: "post",
                body: {
                    src: src,
                    tar: tar,
                }
            }).
                then(response => response.json()).
                then(response => {
                    if (response.err.length > 0) {
                        alert("Error Occured while creating remote tunnel: ", response.err)
                    } else {
                        alert("Successfully Created Local Tunnel at ", response.data.src)
                        tarPort.current.value = ""
                    }
                })

        } else {
            alert("Invalid Port(s)")
        }
    }
    return (
        <div>
            <Header>Local Tunnel</Header>
            <Form>
                <Form.Field>
                    <label>Port to Tunnel</label>
                    <input placeholder="port" ref={tarPort} />
                </Form.Field>
                <Form.Field>
                    <label>Port to be opened</label>
                    <input placeholder="port" ref={srcPort} />
                </Form.Field>
                <Button onClick={handleSubmit}>Create Tunnel</Button>
            </Form>
        </div>
    )
}

export default LocalTunnel;