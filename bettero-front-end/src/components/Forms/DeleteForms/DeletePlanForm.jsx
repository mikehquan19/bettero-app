import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { expenseappClient } from '../../../provider/api';
import handleError from '../../../provider/handleError';
import { faX } from '@fortawesome/free-solid-svg-icons';
import './DeleteFormStyle.css';



// form to delete the plan 
function DeletePlanForm({ intervalType, handleChangePlanList, handleHideForm }) {
  function handleSubmitForm(e) {
    e.preventDefault();
    if (e.target.value === "true") {
      expenseappClient.delete(`budget/${intervalType}`)
        .then((response) => {
          if (response) {
            expenseappClient.get("/budget")
              .then((response) => handleChangePlanList(response.data))
              .catch((error) => handleError(error));
          }
        })
        .catch((planError) => handleError(planError));
    }
    handleHideForm();
  }

  return (
    <>
      <div className="delete-form-wrapper">
        <button className="x-button" onClick={handleHideForm} style={{
          backgroundColor: "skyblue",
          color: "black",
        }}><FontAwesomeIcon icon={faX} /></button>
        <form>
          <label>Do you want to delete the account ?</label>
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

export default DeletePlanForm;