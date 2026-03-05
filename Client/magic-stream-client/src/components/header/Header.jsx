import {useState} from "react";
import Button from "react-bootstrap/Button";
import Container from 'react-bootstrap/Container'
import Nav from "react-bootstrap/Nav"
import Navbar from 'react-bootstrap/Navbar'
import {useNavigate, NavLink, Link} from "react-router-dom";
import useAuth from "../../hook/useAuth.jsx";
import logo from '../../assets/MagicStreamLogo.png';

const Header = ({handleLogout}) =>{
    const navigate = useNavigate();
    //const [auth, setAuth] = useState(false)
    const {auth} = useAuth();

    return(
        <Navbar bg="dark" variant="dark" expand="lg" stick="top" className="shadow-sm">
            <Container>
                <Navbar.Brand>
                    <img
                    alt=""
                    src={logo}
                    width={30}
                    height={30}
                    className="d-inline-block align-top me-2"
                    />
                    Magic Stream
                </Navbar.Brand>

                <Navbar.Toggle area-controls="main-navbar-nav"/>

                <Navbar.Collapse>
                    <Nav className="me-auto">
                        <Nav.Link as = {NavLink} to="/">
                            Home
                        </Nav.Link>
                        <Nav.Link as = {NavLink} to="/recommendedmovies">
                            Recommended
                        </Nav.Link>
                    </Nav>
                    <Nav className = "ms-auto aling-items-center">
                        {auth ? (
                        <>
                            <span className="ms-3 text-light">
                                Hello, <strong>{auth.first_name}   </strong>
                            </span>
                            <Button varient = "outline-light" size="sm" onClick={handleLogout}>
                               Logout
                            </Button>
                        </>
                        ):(
                            <>
                                <Button variant="outline-info" size="sm" className="me-2" onClick={() => navigate("/login")}>
                                    Login
                                </Button>
                                <Button variant="info" size="sm" onClick={() => navigate("/register")}>
                                    Register
                                </Button>
                            </>

                          )}
                    </Nav>
                </Navbar.Collapse>
            </Container>
        </Navbar>
    )
}
export default Header;
