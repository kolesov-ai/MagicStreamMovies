import {useParams} from 'react-router-dom';
import ReactPlayer from 'react-player';
import './StreamMovie.css';

const StreamMovie = () => {

    let params = useParams();
    let key = params.yt_id;

    return (
        <div className="react-player-container">
            {(key!=null)?<ReactPlayer
                slot="media"
                src ={`https://www.youtube.com/watch?v=${key}`}
                controls={true}
                style={{
                    width: "auto%",
                    height: "auto%",
                    aspectRatio: '16/9'
                }}
                />:null}
        </div>
    )
}



export default StreamMovie