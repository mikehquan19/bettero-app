import './TransactionTable.css';
import { useState, useEffect } from 'react';
import { useMediaQuery } from "@uidotdev/usehooks";

// The pagination bar used to switch between pages 
function Pagination({ transactionsPerPage, numTrans, handlePagination }) {
  const [clickedNumber, setClickedNumber] = useState(1);
  const [firstIndex, setFirstIndex] = useState(0);

  useEffect(() => {
    setClickedNumber(firstIndex + 1);
    handlePagination(firstIndex + 1);
  }, [firstIndex]);

  // The array of the numbers in the page
  const pageNumbers = [];
  for (let i = 1; i <= Math.ceil(numTrans / transactionsPerPage); i++) {
    pageNumbers.push(i);
  }

  const filteredNumbers = pageNumbers.slice(firstIndex, firstIndex + 5)

  if (numTrans > transactionsPerPage) {
    return (
      <div className="pagination">
        <button style={{
          color: "darkcyan",
          borderLeft: "0.1rem solid gray",
          borderTopLeftRadius: "12px",
          borderBottomLeftRadius: "12px",
        }} onClick={() => {
          if (firstIndex - 5 > -1) { setFirstIndex(firstIndex - 5) }
        }}>Prev</button>
        {filteredNumbers.map(pageNumber =>
          <button key={pageNumber} onClick={() => {
            setClickedNumber(pageNumber);
            handlePagination(pageNumber);
          }}
            style={clickedNumber === pageNumber ? { backgroundColor: "skyblue", color: "white" } : { backgroundColor: "white", color: "skyblue" }}>{pageNumber}</button>
        )}
        <button style={{
          color: "darkcyan",
          borderRight: "0.1rem solid gray",
          borderTopRightRadius: "12px",
          borderBottomRightRadius: "12px",
        }} onClick={() => {
          if (firstIndex + 5 < pageNumbers.length) { setFirstIndex(firstIndex + 5) }
        }}>Next</button>
      </div>
    );
  }
}

function TransactionTableRow({ transactionInfo }) {
  const [detailRowPresent, setDetailRowPresent] = useState(false);
  return (
    <>
      <tr onClick={() => {
        if (isSmallDevice) {
          setDetailRowPresent(true);
        }
      }}>
        <td>{transactionInfo.occur_date}</td>
        <td>{transactionInfo.description}</td>
        <td>{transactionInfo.category}</td>
        <td>{transactionInfo.account_name}</td>
        <td>${transactionInfo.amount}</td>
      </tr>

      {/* Row displaying the detail */}
      {(isSmallDevice && detailRowPresent) && (
        <tr>
          <td colSpan={3}>
            <div className="transaction-detail">
              <h4>Transaction detail:</h4>
              <div id="first">
                <p><span>Date:</span> {transactionInfo.occur_date}</p>
                <p><span>Description:</span> {transactionInfo.description}</p>
                <p><span>Category:</span> {transactionInfo.category}</p>
                <p><span>Account:</span> {transactionInfo.account_name}</p>
                <p><span>Amount:</span> ${transactionInfo.amount}</p>
              </div>
              <button onClick={() => setDetailRowPresent(false)}>Collapse</button>
            </div>
          </td>
        </tr>
      )}
    </>
  );
}

export default function TransactionTable({ listName, transactionList, transactionCount }) {
  const isSmallDevice = useMediaQuery("only screen and (max-width: 500px)");
  const isMediumDevice = useMediaQuery("only screen and (max-width: 1000px");

  // State of paginations 
  const [currentPage, setCurrentPage] = useState(1);
  const transactionsPerPage = 10;

  return (
    <>
      <div className="latest-transactions">
        <div className="table-caption" >
          <h3 style={{ display: "inline-block" }}>{listName} ({transactionCount}):</h3>
        </div>
        
        {/* the table of transaction's list */}
        <table>
          <thead>
            <tr>
              <th scope="col">Date</th>
              <th scope="col">Description</th>
              <th scope="col">Category</th>
              <th scope="col">Account</th>
              <th scope="col">Amount</th>
            </tr>
          </thead>
          <tbody>
            {transactionList.length === 0 && (
              <tr><td colSpan={5} style={{ fontWeight: "bold" }}>There are no transactions</td></tr>
            )}
            {/* List of transactions. If there are no trasactions, it won't render */}
            {transactionList.map(transaction =>
              <TransactionTableRow
                key={transaction.id} transactionInfo={transaction} 
                isMediumDevice={isMediumDevice} isSmallDevice={isSmallDevice} />
            )}
          </tbody>
        </table>
      </div>
      <Pagination 
        transactionsPerPage={transactionsPerPage} numTrans={transactionCount}
        handlePagination={(pageNumber) => setCurrentPage(pageNumber)} />
    </>
  );
}