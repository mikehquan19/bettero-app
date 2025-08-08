import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { expenseappClient } from '@provider/api';
import handleError from '@provider/handleError';
import { faX } from '@fortawesome/free-solid-svg-icons';
import './DeleteFormStyle.scss';
 
interface DeletePlanProps {
  intervalType: string, 
  onChangePlanList: (data: any) => void, 
  onHide: () => void
}

/**
 * Form to delete the budget plan 
 * @param {DeletePlanProps}
 * @returns {JSX.Element}
 */
function DeletePlanForm({ intervalType, onChangePlanList, onHide } : DeletePlanProps): JSX.Element {

  /**
   * Handle the action to submit the form 
   * @param {any} e 
   * @returns {void}
   */
  function handleSubmitForm(e: any): void {
    e.preventDefault();
    if (e.target.value === "true") {
      expenseappClient.delete(`budget/${intervalType}`)
        .then((response) => {
          if (response) {
            expenseappClient.get("/budget")
              .then((response) => onChangePlanList(response.data))
              .catch((error) => handleError(error));
          }
        })
        .catch((planError) => handleError(planError));
    }
    onHide();
  }

  return (
    <>
      <div className="delete-form-wrapper">
        <button className="x-button" onClick={onHide} style={{
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