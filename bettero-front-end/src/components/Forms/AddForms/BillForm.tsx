import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { faX } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { getBills, getUserAccounts, postOrPutBill } from "@provider/api";
import handleError from "@provider/handleError";
import { Account, Bill } from "@interface";
import { ADD_BILL_VALIDATION } from "./Validation";
import './Form.scss';

interface BillFormProps {
  type: 'ADD' | 'UPDATE', 
  currentData?: Bill, 
  onChangeBillList?: (data: Bill[]) => void, 
  onHide: () => void; 
}

// the form to add bill component 
function BillForm({ type = "ADD", currentData, onChangeBillList, onHide }: BillFormProps) {

  const { register, reset, handleSubmit, formState: { errors } } = useForm();
  const [payAccountList, setPayAccountList] = useState<Account[]>([]);

  /**
   * Fetch the user's list of accounts to pay the bills 
   * @returns {void}
   */
  function fetchPayAccountList(): void {
    getUserAccounts()
      .then((response) => setPayAccountList(response.data))
      .catch((error) => handleError(error));
  }

  /**
   * Pass the existing form data, if any
   * @returns {void}
   */
  function fetchFormData(): void {
    if (!!currentData) {
      reset({ 
        ...currentData, 
        dueDate: currentData.dueDate.toISOString().split('T')[0] 
      });
    }
  }

  // Load all necessary data 
  useEffect(() => fetchPayAccountList(), []);
  useEffect(() => fetchFormData(), [payAccountList]);

  function handleSubmitButton(data: Object): void {
    if (type === "ADD") {
      postOrPutBill(type, data)
        .then((response) => !!onChangeBillList ? onChangeBillList(response.data) : null)
        .catch((error) => handleError(error));
    }
    else if (type === "UPDATE") {
      postOrPutBill(type, data)
        .then((response) => {
          if (response) {
            getBills()
              .then((response) => !!onChangeBillList ? onChangeBillList(response.data) : null)
              .catch((error) => handleError(error));
          }
        })
        .catch((error) => handleError(error));

    }
    onHide();
  }

  const categoryArr = ["Housing", "Automobile", "Medical", "Subscription", "Grocery", "Dining", "Shopping", "Gas", "Others"]

  return (
    <>
      <div className="form-wrapper">
        <h3>{type} BILL</h3>
        <form id="add-bill" onSubmit={handleSubmit(handleSubmitButton)}>
          <div className="form-field">
            <label htmlFor="description">Description:</label>
            <small className="error-message">
              {errors?.description && errors?.description?.message as string}
            </small>
            <input type="text" placeholder="Description" id="description" 
            {...register("description", ADD_BILL_VALIDATION.description)} />
          </div>

          <div className="form-field">
            <label htmlFor="amount">Amount:</label>
            <small className="error-message">
              {errors?.amount && errors?.amount?.message as string}
            </small>
            <input type="number" step="any" placeholder="Amount" id="amount" 
            {...register("amount", ADD_BILL_VALIDATION.amount)} />
          </div>

          <div className="form-field">
            <label htmlFor="category">Category: </label>
            <select {...register("category")}>
              {categoryArr.map(category => <option key={category} value={category}>{category}</option>)}
            </select>
          </div>

          <div className="form-field">
            <label htmlFor="pay_account">Pay account:</label>
            <select {...register("payAccount")}>
              {payAccountList.map(payAccount =>
                <option key={payAccount.id} value={payAccount.id}>{payAccount.name}</option>
              )}
            </select> 
          </div>

          <div className="form-field">
            <label htmlFor="due_date">Due date</label>
            <input type="date" id="dueDate" {...register("dueDate")} />
          </div>
        </form>

        <div className="button-area">
          <button className="submit-button" form="add-bill">Submit</button>
          <button className="close-button" onClick={onHide}>Close</button>
        </div>
        <button className="x-button" onClick={onHide}>
          <FontAwesomeIcon icon={faX} />
        </button>
      </div>
      <div className="overlay"></div>
    </>
  );
}

export default BillForm;