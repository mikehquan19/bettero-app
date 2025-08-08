import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { postOrPutStock } from "@provider/api";
import handleError from "@provider/handleError";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faX } from "@fortawesome/free-solid-svg-icons";
import { ADD_STOCK_VALIDATTION } from "./Validation";
import './FormStyle.css';
import { Stock } from "@interface";

interface StockFormProps {
  type: 'ADD' | 'UPDATE', 
  currentData?: Stock, 
  onChangeStockList?: (data: Stock[]) => void; 
  onChangeStockInfo?: (data: Stock) => void,
  onHide: () => void
}

function StockForm({ type = "ADD", currentData, onChangeStockList, onChangeStockInfo, onHide}: StockFormProps) {
  const { register, handleSubmit, formState: { errors }, reset } = useForm();
  const isFieldDisabled = type !== "ADD";
  const [symbolExisting, setSymbolExisting] = useState<boolean>(false);

  // set some initial data for the form 
  function fetchFormData(): void {
    if (!!currentData) {
      reset({ ...currentData });
    }
  }
  useEffect(fetchFormData, []);

  function handleSubmitButton(data: any): void {
    if (type === "ADD") {
      postOrPutStock(type, data)
        .then((response) => {
          // Set the data about the user's list of stock, and hide the form 
          if (!!onChangeStockList) onChangeStockList(response.data); 
          setSymbolExisting(false); 
          onHide();
        })
        .catch((error) => {
          // If there is no stock with the given symbol, displays error
          if (error.response && error.response.status === 404) {
            setSymbolExisting(true);
          }
          handleError(error);
        });
    } 
    else if (type === "UPDATE") {
      data.lastUpdatedDate = data.lastUpdatedDate.toISOString().split('T')[0];
      postOrPutStock(type, data)
        .then((response) => {
          // set the data about the new list of stock
          if (!!onChangeStockInfo) onChangeStockInfo(response.data.stock);
          onHide();
        })
        .catch((error) => handleError(error));
    }
  }

  return (
    <>
      <div className="form-wrapper">
        <button className="x-button" onClick={onHide} style={{
          backgroundColor: "skyblue",
          color: "black",
        }}><FontAwesomeIcon icon={faX} /></button>
        <h3>{type} STOCKS</h3>
        <p className="error-message" style={{ textAlign: "center" }}>
          {symbolExisting && "*This symbol doesn't exist or overlaps the previous ones"}
        </p>
        <form id="add-stock" onSubmit={handleSubmit(handleSubmitButton)}>
          <div className="form-field">
            <label htmlFor="corporation">Corporation:</label>
            <small className="error-message">
              {errors?.corporation && errors?.corporation?.message as string}
            </small>
            <input type="text" id="corporation" placeholder="Corporation"
              {...register("corporation", ADD_STOCK_VALIDATTION.corporation)} />
          </div>

          <div className="form-field">
            <label htmlFor="name">Name: </label>
            <small className="error-message">
              {errors?.name && errors?.name?.message as string}
            </small>
            <input type="text" id="name" placeholder="Name"
              {...register("name", ADD_STOCK_VALIDATTION.name)} />
          </div>

          <div className="form-field">
            <label htmlFor="symbol">Symbol: </label>
            <small className="error-message">
              {errors?.symbol && errors?.symbol?.message as string}
            </small>
            <input type="text" id="symbol" placeholder="Symbol" disabled={isFieldDisabled}
              {...register("symbol", ADD_STOCK_VALIDATTION.symbol)} />
          </div>

          <div className="form-field">
            <label htmlFor="shares">Shares: </label>
            <small className="error-message">
              {errors?.shares && errors?.shares?.message as string}
            </small>
            <input type="number" id="shares" placeholder="Number of Shares"
              step="any" {...register("shares", ADD_STOCK_VALIDATTION.shares)} />
          </div>
          
        </form>
        <div className="button-area">
          <button type="submit" form="add-stock">Submit</button>
          <button onClick={onHide}>Close</button>
        </div>
      </div>
      <div className="overlay"></div>
    </>
  );
}

export default StockForm; 