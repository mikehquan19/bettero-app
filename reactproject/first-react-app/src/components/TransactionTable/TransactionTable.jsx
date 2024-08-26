import './TransactionTable.css'; 
import { useState } from 'react';
import { useMediaQuery } from "@uidotdev/usehooks";

const TransactionTableRow = ({transactionInfo, isSmallDevice, isMediumDevice}) => {
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

const TransactionTable = ({ listName, transactionList, handleSearchClick = null, tableType = "" }) => {
    const [searchTransName, setSearchTransName] = useState(""); 
    
    // media query 
    const isSmallDevice = useMediaQuery("only screen and (max-width : 500px)");
    const isMediumDevice = useMediaQuery("only screen and (max-width: 1000px");

    // change to value to the info entered in search bar 
    const handleValueChange = (e) => {
        setSearchTransName(e.target.value); 
    } 

    // handle the click of the search button 
    const handleSearchSubmit = (e) => {
        e.preventDefault();
        // callback to modify the transaction list which will be passed in rerendering
        handleSearchClick(searchTransName, tableType); 
        setSearchTransName("");
    }

    return (
        <div className="latest-transactions">
            {/* the title of the table of transaction list */}
            <div className="table-caption" >
                <h3 style={{
                    display: "inline-block",
                }}>{listName} ({transactionList.length})</h3>
                {!isSmallDevice && (
                    <form onSubmit={handleSearchSubmit}>
                        <input
                            value={searchTransName}
                            onChange={handleValueChange}
                            type="search" id="transaction-search-bar" placeholder="Search transaction" />
                        <button
                            id="transaction-search-button"
                            htmlFor="transaction-search-bar">Search</button>
                    </form>
                )}
            </div>
            {/* the table of transaction's list */}
            <table>
                <thead>
                    {/* the list of titles in the table  */}
                    <tr>
                        <th scope="col">Date</th>
                        <th scope="col">Description</th>
                        {!isSmallDevice && <th scope="col">Category</th>}
                        {!isMediumDevice &&  <th scope="col">Account</th>}
                        <th scope="col">Amount</th>
                    </tr>
                </thead>
                <tbody>
                    {
                        // rendering message when there are no transactions 
                        transactionList.length === 0 && (
                            <tr><td colSpan={5} style={{
                                fontWeight: "bold"
                            }}>There are no transactions</td></tr>
                        )
                    }
                    {/* the list of transactions, if there are no trasactions, it won't render anything*/}
                    {transactionList.map(transaction =>
                        <TransactionTableRow 
                            key={transaction.id}
                            transactionInfo={transaction}
                            isMediumDevice={isMediumDevice}
                            isSmallDevice={isSmallDevice} /> 
                    )}
                </tbody>
            </table>
        </div>
    );
}

export default TransactionTable;