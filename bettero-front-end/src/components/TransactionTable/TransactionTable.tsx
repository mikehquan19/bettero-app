import { useState, useEffect, CSSProperties } from 'react';
import { useMediaQuery } from "@uidotdev/usehooks";
import { Transaction } from '@interface';
import './TransactionTable.scss';

interface PaginationProps {
  currentIndex: number, 
  pageSize: number, 
  numRecords: number, 
  onNextPage: (index: number) => void,
}

/**
 * The pagination bar used to switch between pages 
 * @param {PaginationProps} props
 * @returns {JSX.Element | undefined}
 */
function Pagination(
  { currentIndex, pageSize, numRecords, onNextPage }: PaginationProps
): JSX.Element | undefined {
  const [selectedIndex, setSelectedIndex] = useState<number>(currentIndex);

  // First index in the pagination bar 
  const firstIndex = Math.floor((selectedIndex - 1) / 5) * 5;
  // The array of page indices
  const numPages = [];
  for (let i = 1; i <= Math.ceil(numRecords / pageSize); i++) {
    numPages.push(i);
  }
  const numPresentPages = numPages.slice(firstIndex, firstIndex + 5);

  // Selected index will change along with the current index
  // List of transactions already changed with current index
  useEffect(() => setSelectedIndex(currentIndex), [currentIndex]);

  /**
   * Handle the action to select the other page of the bar
   * @param {number} pageIndex 
   * @return {void}
   */
  function handleNextPage(pageIndex: number): void {
    setSelectedIndex(pageIndex); 
    onNextPage(pageIndex);
  }

  if (numRecords > pageSize) {
    return (
      <div className="pagination">
        <button className="left-pagination-button"
          onClick={() => {
            if (firstIndex - 5 >= 0) handleNextPage(firstIndex - 4);
          }}
        >Prev</button>

        {numPresentPages.map(presentPageIndex =>
          <button key={presentPageIndex} onClick={() => handleNextPage(presentPageIndex)}
            style={selectedIndex === presentPageIndex ? 
              { backgroundColor: "skyblue", color: "white" } : 
              { backgroundColor: "white", color: "skyblue" }
            }
          >{presentPageIndex}</button>
        )}

        <button className="right-pagination-button"
          onClick={() => {
            if (firstIndex + 5 <= numPages.length - 1) handleNextPage(firstIndex + 6);
          }}
        >Next</button>
      </div>
    );
  }
}

interface TransactionTableRowProps {
  transactionInfo: Transaction,
  isSmallDevice: boolean, 
  isMediumDevice: boolean, 
}

/**
 * Each row of the transaction table 
 * @param {TransactionTableRowProps} props
 * @returns {JSX.Element}
 */
function TransactionTableRow(
  { transactionInfo, isSmallDevice, isMediumDevice }: TransactionTableRowProps
): JSX.Element {
  const [detailRowPresent, setDetailRowPresent] = useState<boolean>(false);

  return (
    <>
      <tr onClick={() => { isSmallDevice ? setDetailRowPresent(true) : null }}>
        <td>{transactionInfo.occurDate.toString()}</td>
        <td>{transactionInfo.description}</td>
        {!(isSmallDevice || isMediumDevice) && <td>{transactionInfo.category}</td>}
        {!isSmallDevice && <td>{transactionInfo.accountName}</td>}
        <td>${transactionInfo.amount}</td>
      </tr>

      {/* Row displaying the detail for mobile device */}
      {(isSmallDevice && detailRowPresent) && (
        <tr>
          <td colSpan={3}>
            <div className="transaction-detail">
              <h3>Transaction detail:</h3>
              <div className="first">
                <p><span>Date:</span> {transactionInfo.occurDate.toString()}</p>
                <p><span>Description:</span> {transactionInfo.description}</p>
                <p><span>Category:</span> {transactionInfo.category}</p>
                <p><span>Account:</span> {transactionInfo.accountName}</p>
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

interface TransactionTableProps {
  listName: string, 
  currentIndex: number,
  transactionList: Transaction[], 
  transactionCount: number,
  onChangeTransList?: (pageIndex: number) => void, 
}

/**
 * The table showing list of transactions 
 * @param {TransactionTableProps} props
 * @returns {JSX.Element}
 */
export default function TransactionTable(
  { listName, currentIndex, transactionList, transactionCount, onChangeTransList }: TransactionTableProps
): JSX.Element {
  const isSmallDevice = useMediaQuery("only screen and (max-width: 700px)");
  const isMediumDevice = useMediaQuery("only screen and (min-width: 700px) and (max-width: 1200px)");
  const transactionsPerPage = 10;

  return (
    <>
      <div className="latest-transactions">
        <div className="table-caption" >
          <h3 style={{ display: "inline-block" }}>{listName} ({transactionCount}):</h3>
        </div>
        <table>
          <thead>
            <tr>
              <th scope="col">Date</th>
              <th scope="col">Description</th>
              {!(isSmallDevice || isMediumDevice) && <th scope="col">Category</th>}
              {!isSmallDevice && <th scope="col">Account</th>}
              <th scope="col">Amount</th>
            </tr>
          </thead>
          <tbody>
            {!transactionList.length && 
              <tr>
                <td colSpan={5} style={{ fontWeight: "bold" }}>There are no transactions</td>
              </tr>
            }
            {transactionList.map(transaction =>
              /* List of transactions. If there are no transactions, it won't render */
              <TransactionTableRow
                key={transaction.id} transactionInfo={transaction}  
                isMediumDevice={isMediumDevice} isSmallDevice={isSmallDevice} />
            )}
          </tbody>
        </table>
      </div>
      <Pagination
        currentIndex={currentIndex}
        pageSize={transactionsPerPage} numRecords={transactionCount}
        onNextPage={(pageIndex: number) => {
          if(!!onChangeTransList) onChangeTransList(pageIndex);
        }} />
    </>
  );
}