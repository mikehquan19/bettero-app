import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faCirclePlus } from '@fortawesome/free-solid-svg-icons';
import './Button.css';

// the Button component
export default function Button({ content, id = null, handleEvents = null }) {
  return (
    <div style={{ textAlign: "center" }}>
      <button className="add-button" id={id}
        onMouseOver={(e) => {
          e.target.style.backgroundColor = "white";
          e.target.style.border = "2px solid darkcyan";
          e.target.style.color = "darkcyan";
        }}
        onMouseOut={(e) => {
          e.target.style.backgroundColor = "darkcyan";
          e.target.style.border = "1px solid green";
          e.target.style.color = "white";
        }}
        onClick={handleEvents}><FontAwesomeIcon icon={faCirclePlus} style={{
          marginRight: "5px",
          display: "block",
          margin: "0 auto 5px auto",
        }} /> {content} </button>
    </div>
  );
}