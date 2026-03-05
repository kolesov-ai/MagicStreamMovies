import {useLocation, Navigate, Outlet} from "react-router-dom";
import useAuth from "../hook/useAuth.jsx";
import Spinner from "./spinner/Spinner.jsx";

const RequiredAuth = () => {
    const {auth, loading} = useAuth();
    const location = useLocation();

    if (loading) {
        return (<Spinner/>)
    }

    return auth ? (
        <Outlet/>
    ) : (
        <Navigate to = '/login' state={{from:location}} replace></Navigate>
    );
};

export default RequiredAuth;