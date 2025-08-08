import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { faX } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { expenseappClient, getUserAccounts } from "@provider/api";
import handleError from "@provider/handleError";
import { Account, Bill } from "@interface";
import { ADD_BILL_VALIDATION } from "./Validation";
import './FormStyle.css';

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


  function fetchFormData(): void {
    if (!!currentData) {
      // pass the existing form data
      reset({ ...currentData });
    }
  }

  // function to fetch the data about the account list
  function fetchPayAccountList(): void {
    getUserAccounts()
      .then((response) => setPayAccountList(response.data))
      .catch((error) => handleError(error));
  }

  // load all necessary data 
  useEffect(() => {
    fetchFormData();
    fetchPayAccountList();
  }, []);

  function handleSubmitButton(data: Object): void {
    if (type === "ADD") {
      expenseappClient.post("/bills", data)
        .then((response) => !!onChangeBillList ? onChangeBillList(response.data) : null)
        .catch((error) => handleError(error));
    }
    else if (type === "UPDATE") {
      expenseappClient.put(`/bills/${currentData!.id}`, data)
        .then((response) => {
          if (response) {
            expenseappClient.get(`/bills`)
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
              {categoryArr.map(category => <option id={category} value={category}>{category}</option>)}
            </select>
          </div>

          <div className="form-field">
            <label htmlFor="pay_account">Pay account:</label>
            <select id="pay_account" {...register("pay_account")}>
              {payAccountList.map(payAccount =>
                <option key={payAccount.id} value={payAccount.id}>{payAccount.name}</option>
              )}
            </select> 
          </div>

          <div className="form-field">
            <label htmlFor="due_date">Due date</label>
            <input type="date" id="due_date" {...register("due_date")} />
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