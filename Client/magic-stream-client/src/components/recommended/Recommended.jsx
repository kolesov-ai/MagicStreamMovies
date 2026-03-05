import useAxiosPrivate from "../../hook/useAxiosPrivate";
import {useEffect, useState} from "react";
import Movies from "../movies/Movies.jsx";

const Recommended = () => {
    const [movies, setMovies] = useState([]);
    const [loading, setLoading] = useState(false);
    const [message, setMessage] =useState();
    const axiosPrivate = useAxiosPrivate();

    useEffect(() => {
        const fetchRecommendedMovies = async () => {
            setLoading(true);
            setMessage("");

            console.log("ST0:")

            try{
                console.log("ST1:")
                const response = await axiosPrivate.get('/recommendedmovies');
                console.log("ST2:")
                setMovies(response.data);
                console.log("ST3:")
            } catch (error){
                console.error("Error fetching recommended movies:", error)
            } finally {
                setLoading(false);
            }

        }
        fetchRecommendedMovies();
    }, [])

    return (
        <>
            {loading ? (
                <h2>Loading...</h2>
            ) :(
                <Movies movies = {movies} message ={message} />
            )}
        </>
    )
}

export default Recommended