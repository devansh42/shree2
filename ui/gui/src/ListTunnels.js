import React, { useState, useEffect } from "react";
import { Button, Header, Table } from "semantic-ui-react";
import backendServer from "./const";


function ListTunnels(props) {
    const [listContents, setListContents] = useState([])
    useEffect(() => {
        fetch(backendServer.concat("/tunnels"), {
            method: "get",
        }).
            then(response => response.json()).
            then(response => {
                if (response.err.length > 0) {
                    alert("Error Occured while listing tunnels: ", response.err)
                } else {
                    setListContents(response.data.tunnels)
                }
            })
    }, [1])

    const removeRow = (index) => {
        const val = listContents[index]
        fetch(backendServer.concat("/tunnel"), {
            method: "delete",
            body: {
                isRemote: val.isRemote,
                port: val.tar,
            }
        }).
            then(response => response.json()).
            then(response => {
                if (response.err.length > 0) {
                    alert("Error Occured while disconnecting tunnel: ", response.err)
                } else {
                    alert("Tunnel removed successfully")
                    const newContents = listContents.filter((v, i) => {
                        return i !== index;
                    })
                    setListContents(newContents);
                }
            })
    }
    const tunnelRow = (tunnel, i) => {
        return <div>
            <Header>List of Tunnel(s)</Header>
            <Table.Row key={i}>
                <Table.Cell>
                    {tunnel.isRemote ? "Remote" : "Local"}
                </Table.Cell>
                <Table.Cell>
                    {tunnel.tar}
                </Table.Cell>
                <Table.Cell>
                    {tunnel.src}
                </Table.Cell>
                <Table.Cell>
                    <Button negative onClick={removeRow(i)}>Disconnect</Button>
                </Table.Cell>
            </Table.Row></div >
    };
    return (
        <Table>
            <Table.Header>
                <Table.Row>
                    <Table.HeaderCell>
                        Type
                    </Table.HeaderCell>
                    <Table.HeaderCell>
                        Actual Port
                    </Table.HeaderCell>
                    <Table.HeaderCell>
                        Tunneled Port
                    </Table.HeaderCell>
                </Table.Row>
            </Table.Header>
            <Table.Body>
                {listContents.map(tunnelRow)}
            </Table.Body>
        </Table>
    )
}

export default ListTunnels;