import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { useNavigate } from "react-router-dom";
import { faX } from '@fortawesome/free-solid-svg-icons';
import './DeleteFormStyle.css';

function LogoutForm({ onHideForm }) {
  const navigate = useNavigate();

  function handleSubmitForm(e) {
    e.preventDefault();
    if (e.target.value === "true") navigate("/logout");
    onHideForm();
  }

  return (
    <>
      <div className="delete-form-wrapper">
        <button className="x-button" onClick={onHideForm} style={{
          backgroundColor: "skyblue",
          color: "black"
        }}><FontAwesomeIcon icon={faX} /></button>
        <form style={{
          marginTop: "1em",
        }}>
          <label>Do you want to log out ?</label>
          <div className="delete-form-input">
            <button onClick={handleSubmitForm} type="submit" value="true">YES</button>
            <button onClick={handleSubmitForm} type="submit" value="false">NO</button>
          </div>
        </form>
      </div>
      <div className="overlay"></div>
    </>
  );
}

export default LogoutForm;