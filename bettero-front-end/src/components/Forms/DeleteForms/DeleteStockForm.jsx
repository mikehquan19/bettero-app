import { expenseappClient } from "../../../provider/api";
import handleError from "../../../provider/handleError";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faX } from '@fortawesome/free-solid-svg-icons';
import './DeleteFormStyle.css';

function DeleteStockForm({stockSymbol, onChangeStockList, onHideForm }) {

  function handleSubmitForm(e) {
    e.preventDefault();
    if (e.target.value === "true") {
      expenseappClient.delete(`/stocks/${stockSymbol}`)
        .then((response) => {
          if (response) {
            expenseappClient.get(`/stocks`)
              .then((response) => onChangeStockList(response.data))
              .catch((error) => handleError(error));
          }
        })
        .catch((stockError) => handleError(stockError));
    }
    onHideForm();
  }

  return (
    <>
      <div className="delete-form-wrapper">
        <button className="x-button" onClick={onHideForm} style={{
          backgroundColor: "skyblue",
          color: "black",
        }}><FontAwesomeIcon icon={faX} /></button>
        <form style={{ marginTop: "1em" }}>
          <label>Do you want to delete this stock ?</label>
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

export default DeleteStockForm;