import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { faX } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { expenseappClient } from "../../../provider/api";
import handleError from "../../../provider/handleError";
import './FormStyle.css';


// the form to add bill component 
const BillForm = ({
  type = "ADD",
  currentFormData = null,
  onChangeBillList = null,
  onHideForm,
}) => {

  // necessary components of useForm() hook 
  const { register, reset, handleSubmit, formState: { errors } } = useForm();
  const [payAccountList, setPayAccountList] = useState([]);

  // function to load the form data if any
  const fetchFormData = () => {
    if (currentFormData !== null) {
      // pass the existing form data along with the userId
      reset({ ...currentFormData, });
    }
  }

  // function to fetch the data about the account list
  const fetchPayAccountList = () => {
    expenseappClient.get("/accounts")
      .then((response) => {
        setPayAccountList(response.data);
      })
      .catch((error) => handleError(error));
  }

  // load all necessary data 
  useEffect(() => {
    fetchFormData();
    fetchPayAccountList();
  }, []);

  const handleSubmitButton = (data, e) => {
    e.preventDefault();

    if (type === "ADD") {
      expenseappClient.post("/bills", data)
        .then((response) => onChangeBillList(response.data))
        .catch((error) => handleError(error));
    }
    else if (type === "UPDATE") {
      expenseappClient.put(`/bills/${currentFormData.id}`, data)
        .then((response) => {
          if (response) {
            expenseappClient.get(`/bills`)
              .then((response) => onChangeBillList(response.data))
              .catch((error) => handleError(error));
          }
        })
        .catch((error) => handleError(error));

    }
    onHideForm();
  }

  // validation of each field in the form 
  const fieldValidation = {
    description: {
      required: "*Description is required",
    },
    amount: {
      required: "*Amount is required",
      min: {
        value: 0,
        message: "*Balance must be greater than 0",
      },
    },
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
              {errors?.description && errors.description.message}
            </small>
            <input type="text" placeholder="Description" name="description"
              id="description" {...register("description", fieldValidation.description)} />
          </div>

          <div className="form-field">
            <label htmlFor="amount">Amount:</label>
            <small className="error-message">
              {errors?.amount && errors.amount.message}
            </small>
            <input type="number" step="any" placeholder="Amount" name="amount"
              id="amount" {...register("amount", fieldValidation.amount)} />
          </div>

          <div className="form-field">
            <label htmlFor="category">Category: </label>
            <select name="category" {...register("category")}>
              {categoryArr.map(category => 
                  <option id={category} value={category}>{category}</option>
              )}
            </select>
          </div>

          <div className="form-field">
            <label htmlFor="pay_account">Pay account:</label>
            <select name="pay_account" id="pay_account" {...register("pay_account")}>
              {payAccountList.map(payAccount =>
                <option key={payAccount.id} value={payAccount.id}>{payAccount.name}</option>
              )}
            </select>
          </div>

          <div className="form-field">
            <label htmlFor="due_date">Due date</label>
            <input type="date" id="due_date" name="due_date" {...register("due_date")} />
          </div>
        </form>

        <div className="button-area">
          <button className="submit-button" form="add-bill">Submit</button>
          <button className="close-button" onClick={onHideForm}>Close</button>
        </div>
        <button className="x-button" onClick={onHideForm}>
          <FontAwesomeIcon icon={faX} />
        </button>
      </div>
      <div className="overlay"></div>
    </>
  );
}

export default BillForm;