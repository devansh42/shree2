import { Container, Menu } from 'semantic-ui-react';
import {
  BrowserRouter as Router,
  Link,
  Routes,
  Route,
} from "react-router-dom";
import ListTunnels from './ListTunnels';
import LocalTunnel from './LocalTunnel';
import RemoteTunnel from './RemoteTunnel';

function App() {
  return (
    <div>
      <Container>
        <Router>
          <Menu>
            <Menu.Item active={true}>
              Shree
          </Menu.Item>
            <Menu.Item >
              <Link to="/localTunnel">
                Local Tunnel
            </Link>
            </Menu.Item>
            <Menu.Item >
              <Link to="/remoteTunnel">
                Remote Tunnel
            </Link>
            </Menu.Item>
            <Menu.Item >
              <Link to="/list">
                List Tunnel
            </Link>
            </Menu.Item>

          </Menu>
          <div>
            <Routes>
              <Route element={<ListTunnels />} path="/">
              </Route>
              <Route element={<ListTunnels />} path="/list">
              </Route>
              <Route element={<LocalTunnel />} path="/localTunnel">
              </Route>
              <Route element={<RemoteTunnel />} path="/remoteTunnel">
              </Route>
            </Routes>
          </div>
        </Router>
      </Container>
    </div>
  );
}

export default App;
