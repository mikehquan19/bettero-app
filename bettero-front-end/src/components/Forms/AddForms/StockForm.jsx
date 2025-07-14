import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { expenseappClient } from "../../../provider/api";
import handleError from "../../../provider/handleError";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faX } from "@fortawesome/free-solid-svg-icons";
import { ADD_STOCK_VALIDATTION } from "./Validation";
import './FormStyle.css';

const StockForm = ({
  type = "ADD", currentFormData = null, 
  onChangeStockList = null, onChangeStockInfo = null, onHideForm,
}) => {
  // necessary components of useForm() hook 
  const { register, handleSubmit, formState: { errors }, reset } = useForm();
  const isFieldDisabled = type === "ADD" ? false : true;
  const [symbolExisting, setSymbolExisting] = useState(false);

  // set some initial data for the form 
  const fetchFormData = () => {
    if (currentFormData != null) {
      reset({ ...currentFormData });
    } else {

    }
  }
  useEffect(fetchFormData, []);

  const handleSubmitButton = (data, e) => {
    e.preventDefault();
    // console.log data for checking 
    console.log(data);

    // add new stock
    if (type === "ADD") {
      // call the POST method API to add new stocks
      expenseappClient.post(`/stocks`, data)
        .then((response) => {
          onChangeStockList(response.data); // set the data about the user's list of stock
          setSymbolExisting(false); // hide the form 
          onHideForm();
        })
        .catch((error) => {
          // if there is no stock with the given symbol, displays error
          if (error.response && error.response.status === 404) {
            setSymbolExisting(true);
          }
          handleError(error);
        });
    }
    // update current stock
    else if (type === "UPDATE") {
      // call the PUT method API to update current stock 
      expenseappClient.put(`/stocks/${data.symbol}`, data)
        .then((response) => {
          // set the data about the new list of stock
          onChangeStockInfo(response.data.stock);
          onHideForm();
        })
        .catch((error) => handleError(error));
    }
  }

  return (
    <>
      <div className="form-wrapper">
        <button className="x-button" onClick={onHideForm} style={{
          backgroundColor: "skyblue",
          color: "black",
        }}><FontAwesomeIcon icon={faX} /></button>
        <h3>{type} STOCKS</h3>
        {type === "UPDATE" && (
          <p style={{
            textAlign: "center",
            color: "red",
          }}>You can't update the symbol of the stock</p>
        )}
        <p className="error-message" style={{
          textAlign: "center",
        }}>{symbolExisting && "*This symbol doesn't exist or overlaps the previous ones"}</p>
        <form id="add-stock" onSubmit={handleSubmit(handleSubmitButton)}>
          <div className="form-field">
            <label htmlFor="corporation">Corporation:</label>
            <small className="error-message">
              {errors?.corporation && errors.corporation.message}
            </small>
            <input type="text" id="corporation" placeholder="Corporation"
              name="corporation" {...register("corporation", ADD_STOCK_VALIDATTION.corporation)} />
          </div>

          <div className="form-field">
            <label htmlFor="name">Name: </label>
            <small className="error-message">
              {errors?.name && errors.name.message}
            </small>
            <input type="text" id="name" placeholder="Name"
              name="name" {...register("name", ADD_STOCK_VALIDATTION.name)} />
          </div>

          <div className="form-field">
            <label htmlFor="symbol">Symbol: </label>
            <small className="error-message">
              {errors?.symbol && errors.symbol.message}
            </small>
            <input type="text" id="symbol" placeholder="Symbol" disabled={isFieldDisabled}
              name="symbol" {...register("symbol", ADD_STOCK_VALIDATTION.symbol)} />
          </div>

          <div className="form-field">
            <label htmlFor="shares">Shares: </label>
            <small className="error-message">
              {errors?.shares && errors.shares.message}
            </small>
            <input type="number" id="shares" placeholder="Number of Shares"
              name="shares" step="any" {...register("shares", ADD_STOCK_VALIDATTION.shares)} />
          </div>
          
        </form>
        <div className="button-area">
          <button type="submit" form="add-stock">Submit</button>
          <button onClick={onHideForm}>Close</button>
        </div>
      </div>
      <div className="overlay"></div>
    </>
  );
}

export default StockForm; 