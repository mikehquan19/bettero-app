import { useState, useEffect } from 'react';
import { useMediaQuery } from '@uidotdev/usehooks';
import { expenseappClient } from '../../provider/api';
import BillForm from '../Forms/AddForms/BillForm';
import DeleteBillForm from '../Forms/DeleteForms/DeleteBillForm';
import handleError from '../../provider/handleError';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faPlusCircle } from '@fortawesome/free-solid-svg-icons';
import { reformatDate } from '../../utils';
import './BillTable.css'

function BillTableRow({ billInfo, onChangeBillList }) {
  const [billFormPresent, setBillFormPresent] = useState(false);
  const [deleteBillFormPresent, setDeleteBillFormPresent] = useState(false);
  const [detailRowPresent, setDetailRowPresent] = useState(false);
  const isMobileDevice = useMediaQuery("only screen and (max-width : 500px)");
  const isMediumDevice = useMediaQuery("only screen and (max-width : 1000px)");

  return (
    <>
      <tr onClick={() => {
        if (isMobileDevice) setDetailRowPresent(!detailRowPresent);
      }}>
        <td>{billInfo.description}</td>
        {(!isMediumDevice || isMobileDevice) && <td>${billInfo.amount}</td>}
        {(!isMediumDevice && !isMobileDevice) && <td>{billInfo.category}</td>}
        {!isMediumDevice && <td>{billInfo.pay_account_name}</td>}
        <td>{billInfo.due_date}</td>
        {!isMobileDevice && (
          <>
            <td>
              {/* the button to delete the bills */}
              <button className="bill-button"
                onClick={() => setDeleteBillFormPresent(true)}>Delete</button>
            </td>
            <td>
              {/* the button to update the bills */}
              <button className="bill-button"
                onClick={() => setBillFormPresent(true)}>Update</button>
            </td>
          </>
        )}
      </tr>
      {(isMobileDevice && detailRowPresent) && (
        <tr><td colSpan={3}>
          <div className="bill-detail">
            <h3>Bill detail</h3>
            <p><strong>Description:</strong> {billInfo.description}</p>
            <p><strong>Amount:</strong> ${billInfo.amount}</p>
            <p><strong>Categoty:</strong> {billInfo.category}</p>
            <p><strong>Pay account:</strong> {billInfo.pay_account_name}</p>
            <p><strong>Due date:</strong> {billInfo.due_date}</p>

            {/* the button to delete the bills */}
            <div style={{ display: "flex", flexDirection: "row", gap: "5px", justifyContent: "center" }}>
              <button className="bill-button" onClick={() => setDetailRowPresent(false)}>Back</button>
              <button className="bill-button" onClick={() => setDeleteBillFormPresent(true)}>Delete</button>
              <button className="bill-button" onClick={() => setBillFormPresent(true)}>Update</button>
            </div>
          </div>
        </td></tr>
      )}
      
      {billFormPresent && (<BillForm
        type="UPDATE"
        currentFormData={{
          id: billInfo.id,
          description: billInfo.description,
          amount: billInfo.amount,
          category: billInfo.category, 
          pay_account: billInfo.pay_account,
          due_date: reformatDate(billInfo.due_date, '/', '-'),
        }}

        onChangeBillList={onChangeBillList}
        onHideForm={() => setBillFormPresent(false)} />)}

      {deleteBillFormPresent && (<DeleteBillForm
        billId={billInfo.id}
        onChangeBillList={onChangeBillList}
        onHideForm={() => setDeleteBillFormPresent(false)} />)}
    </>
  );
}

// the card showing the information of the bills 
export default function BillTable() {
  
  // the list of bills to be displayed
  const [billList, setBillList] = useState([]);
  const [billFormPresent, setBillFormPresent] = useState(false);
  const isMobileDevice = useMediaQuery("only screen and (max-width : 500px)");
  const isMediumDevice = useMediaQuery("only screen and (max-width : 1000px)");

  // function to fetch the user's list of bills
  function fetchBillList() {
    expenseappClient.get("/bills")
      .then((response) => setBillList(response.data))
      .catch((error) => handleError(error));
  }
  // fetch data only once 
  useEffect(fetchBillList, []);

  return (
    <div className="bills-list">
      {/* The title of the table of bill's list */}
      <div className="table-title">
        <h3 style={{display: "inline-block", margin: "0"}}>Bills this month ({billList.length})</h3>
      </div>
      {/* The table of list of bills */}
      <table>
        <thead>
          <tr>
            <th>Description</th>
            {(!isMediumDevice || isMobileDevice) && <th>Amount</th>}
            {(!isMediumDevice && !isMobileDevice) && <th>Category</th>}
            {!isMediumDevice && <th>Pay account</th>}
            <th>Due date</th>
            {!isMobileDevice && (
              <>
                <th>Delete</th>
                <th>Update</th>
              </>
            )}
          </tr>
        </thead>
        <tbody>
          {billList.length === 0 && (
            <tr><td colSpan={isMobileDevice ? 3 : 7} style={{fontWeight: "bold"}}>You haven't added any bills</td></tr>
          )}
          {billList.map(bill =>
            <BillTableRow key={bill.id} billInfo={bill}
              onChangeBillList={(newBillList) => setBillList(newBillList)} />)}
        </tbody>
      </table>
      {/* The button to add the new bills for this month */}
      <button className="add-bill-button" onClick={() => setBillFormPresent(true)}>
        <FontAwesomeIcon icon={faPlusCircle} style={{ display: "block", margin: "0 auto 5px auto" }} />Add Bill
      </button>
      {/* The form to add bill */}
      {billFormPresent && (<BillForm
        type="ADD"
        onChangeBillList={(newBillList) => setBillList(newBillList)}
        onHideForm={() => setBillFormPresent(false)} />)}
    </div>
  );
}