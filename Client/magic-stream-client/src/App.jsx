import { useState } from 'react'
import './App.css'
import Home from './components/home/Home.jsx';
import Header from './components/header/Header.jsx';
import Register from './components/register/Register';
import Login from "./components/login/Login.jsx";
import {Route, Routes, useNavigate} from "react-router-dom"
import Layout from "./components/Layout.jsx";
import RequiredAuth from "./components/RequiredAuth.jsx";

import Recommended from "./components/recommended/Recommended.jsx";
import Review from "./components/review/Review.jsx";
import useAuth from "./hook/useAuth.jsx";
import axiosClient from "./api/axiosConfig"
import StreamMovie from "./components/stream/StreamMovie.jsx";



function App() {

    const navigate = useNavigate();
    const { auth, setAuth } = useAuth();


    const updateMovieReview = (imdb_id) => {
        navigate(`/updatereview/${imdb_id}`);
    };

    const [count, setCount] = useState(0)

    const handleLogout = async () => {
        try{
            const response = await axiosClient.post("/logout", {user_id: auth.user_id});
            console.log(response.data);
            setAuth(null);
            //localStorage.removeItem('user');
            console.log('User logget out');
        } catch (error) {
            console.error('Error logging out:', error)
        }
    };

  return (
    <>
        <Header handleLogout={handleLogout}/>
        <Routes path = "/" element = {<Layout/>}>
            <Route path = "/" element={<Home updateMovieReview={updateMovieReview}/>}></Route>
            <Route path = "/register" element={<Register/>}></Route>
            <Route path = "/login" element={<Login/>}></Route>
            <Route element={<RequiredAuth/>}>
                <Route path="/recommendedmovies" element={<Recommended/>}></Route>
                <Route path="/updatereview/:imdb_id" element={<Review/>}></Route>
                <Route path="/stream/:yt_id" element={<StreamMovie/>}></Route>
            </Route>
        </Routes>
    </>
  )
}

export default App
