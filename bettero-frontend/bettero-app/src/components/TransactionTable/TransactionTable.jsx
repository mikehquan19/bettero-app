import './TransactionTable.css';
import { useState, useEffect } from 'react';
import { useMediaQuery } from "@uidotdev/usehooks";

// the pagination bar used to switch between pages 
const Pagination = ({ transPerPage, numTrans, handlePagination }) => {
  const [clickedNumber, setClickedNumber] = useState(1);
  const [firstIndex, setFirstIndex] = useState(0);

  useEffect(() => {
    setClickedNumber(firstIndex + 1);
    handlePagination(firstIndex + 1);
  }, [firstIndex])

  const pageNumbers = []; // the array of the numbers in the page
  for (let i = 1; i <= Math.ceil(numTrans / transPerPage); i++) {
    pageNumbers.push(i);
  }

  const filteredNumbers = pageNumbers.slice(firstIndex, firstIndex + 5)

  if (numTrans > transPerPage) {
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

const TransactionTableRow = ({ transactionInfo, isSmallDevice, isMediumDevice }) => {
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
        {!isSmallDevice && <td>{transactionInfo.category}</td>}
        {!isMediumDevice && <td>{transactionInfo.account_name}</td>}
        <td>${transactionInfo.amount}</td>
      </tr>

      {/* The row expressing the detail  */}
      {(isSmallDevice && detailRowPresent) && (
        <tr><td colSpan={3}>
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
        </td></tr>
      )}
    </>
  );
}

const TransactionTable = ({ listName, transactionList }) => {
  // media query 
  const isSmallDevice = useMediaQuery("only screen and (max-width : 500px)");
  const isMediumDevice = useMediaQuery("only screen and (max-width: 1000px");

  // state for paginations 
  const [currentPage, setCurrentPage] = useState(1);
  const transactionsPerPage = 10;

  // set the first and last index of the current page 
  const lastBetIndex = currentPage * transactionsPerPage;
  const firstBetIndex = lastBetIndex - transactionsPerPage;
  const currentTransList = transactionList.slice(firstBetIndex, lastBetIndex)

  return (
    <>
      <div className="latest-transactions">
        {/* the title of the table of transaction list */}
        <div className="table-caption" >
          <h3 style={{ display: "inline-block" }}>{listName} ({transactionList.length})</h3>
        </div>
        
        {/* the table of transaction's list */}
        <table>
          <thead>
            {/* the list of titles in the table  */}
            <tr>
              <th scope="col">Date</th>
              <th scope="col">Description</th>
              {!isSmallDevice && <th scope="col">Category</th>}
              {!isMediumDevice && <th scope="col">Account</th>}
              <th scope="col">Amount</th>
            </tr>
          </thead>
          <tbody>
            { // rendering message when there are no transactions 
              transactionList.length === 0 && (
                <tr><td colSpan={5} style={{ fontWeight: "bold" }}>There are no transactions</td></tr>
              )
            }
            {/* the list of transactions, if there are no trasactions, it won't render anything*/}
            {currentTransList.map(transaction =>
              <TransactionTableRow
                key={transaction.id}
                transactionInfo={transaction} isMediumDevice={isMediumDevice} isSmallDevice={isSmallDevice} />
            )}
          </tbody>
        </table>
      </div>
      <Pagination transPerPage={transactionsPerPage} numTrans={transactionList.length}
        handlePagination={(pageNumber) => setCurrentPage(pageNumber)} />
    </>
  );
}

export default TransactionTable;