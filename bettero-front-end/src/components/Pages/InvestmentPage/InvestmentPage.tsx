import { useState, useEffect } from "react";
import { useMediaQuery } from "@uidotdev/usehooks";
import TimeseriesChart from "@Charts/TimeseriesChart";
import { faPlusCircle } from '@fortawesome/free-solid-svg-icons';
import { getStockPrice, getPorfolioValues, getStocks} from "@provider/api";
import handleError from "@provider/handleError";
import StockForm from "@Forms/AddForms/StockForm";
import DeleteStockForm from "@Forms/DeleteForms/DeleteStockForm";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { reformatDate } from "@utils";
import { Stock } from "@interface";
import './InvestmentPage.css';

interface StockTableRowProps {
  stockInfoData: Stock, 
  onChangeStockList: (data: Stock[]) => void
}

/**
 * The row in the table of stock
 * @param {StockTableRowProps} 
 * @returns {JSX.Element}
 */
function StockTableRow({ stockInfoData, onChangeStockList }: StockTableRowProps): JSX.Element {
  const [stockInfo, setStockInfo] = useState<Stock>({ ...stockInfoData });
  const [detailRowPresent, setDetailRowPresent] = useState(false);

  const [stockPriceObject, setStockPriceObject] = useState<Record<string, number>>({});
  const [loadStockPrice, setLoadStockPrice] = useState(false);

  const [updateStockFormPresent, setUpdateStockFormPresent] = useState(false);
  const [deleteStockFormPresent, setDeleteStockFormPresent] = useState(false);

  // conditional rendering with useMediaQuery 
  const isMobileDevice = useMediaQuery("only screen and (max-width : 500px)");
  const isMediumDevice = useMediaQuery("only screen and (max-width : 1000px)");

  /**
   * Fetch the list of price of the stocks 
   */
  function fetchPriceDetail(): void {
    // if the stock price hasn't been loaded yet, load it
    if (!loadStockPrice) {
      setLoadStockPrice(true);

      getStockPrice(stockInfo.symbol)
        .then((response) => setStockPriceObject(response.data.priceList))
        .catch((error) => handleError(error));
    }
  }

  // compute the value and value change of shares of stock
  const sharesValue = (stockInfo.currentClose * stockInfo.shares).toFixed(2);
  const sharesValueChange = Number((stockInfo.change * stockInfo.shares).toFixed(2));

  return (
    <>
      <tr className="basic-row"
        onClick={() => {
          // responsive if the the device's screen is mobile
          if (isMobileDevice) {
            fetchPriceDetail();
            setDetailRowPresent(!detailRowPresent);
          }
        }}>
        <td>{stockInfo.symbol}</td>
        {!isMediumDevice && <td>{stockInfo.name}</td>}
        <td>
          <span style={{ display: "block" }}>${stockInfo.currentClose}</span> {stockInfo.change > 0 && "+"}{stockInfo.change}
        </td>
        <td>
          <span style={{ display: "block" }}>${sharesValue}</span> {sharesValueChange > 0 && "+"}{sharesValueChange}
        </td>

        {!isMediumDevice && <td>{stockInfo.lastUpdatedDate.toString()}</td>}
        {!isMobileDevice && (
          <>
            {/* the button to see details, update, or delete the stock */}
            <td>
              <button onClick={() => {
                fetchPriceDetail(); // load the stock price first, if not yet
                setDetailRowPresent(!detailRowPresent);
              }}>
                {detailRowPresent ? (<span>Back</span>) : (<span>More</span>)}
              </button>
            </td>
            <td><button onClick={() => setUpdateStockFormPresent(true)}>Update</button></td>
            <td><button onClick={() => setDeleteStockFormPresent(true)}>Delete</button></td>
          </>
        )}
      </tr>

      {detailRowPresent && (
        <tr className="detail-row"><td colSpan={8}>
          <div className="detail-row-wrapper">
            <h3 style={{ textAlign: "center", margin: "20px auto" }}>
              The stock's detail last updated on {stockInfo.lastUpdatedDate.toString()} (updated everyday)
            </h3>
            <div className="grid-stock-wrapper">
              <div className="stock-detail">
                <div>
                  {isMobileDevice && <p><strong>Name: </strong> {stockInfo.name}</p>}
                  <p><strong>Corporation:</strong> {stockInfo.corporation}</p>
                  <p><strong>Shares:</strong> {stockInfo.shares}</p>
                  <p><strong>Open:</strong> ${stockInfo.open}</p>
                  <p><strong>Low:</strong> ${stockInfo.low}</p>
                  <p><strong>High:</strong> ${stockInfo.high}</p>
                  <p><strong>Volume:</strong> {stockInfo.volume}</p>
                </div>
              </div>
              <div className="stock-price-chart">
                <TimeseriesChart intervalType="month" contentType="stock price" timeSeriesObject={stockPriceObject} />
              </div>
            </div>
            <div className="button-area">
              <button className="back-button" onClick={() => setDetailRowPresent(false)}>Back</button>
              {/* if accessed by mobile screen */
                isMobileDevice && (
                  <>
                    <button onClick={() => setUpdateStockFormPresent(true)}>Update</button>
                    <button onClick={() => setDeleteStockFormPresent(true)}>Delete</button>
                  </>
                )
              }
            </div>
          </div>
        </td></tr>
      )}

      { /* The form to update the stock's info */
        updateStockFormPresent && 
          <tr>
            <td>
              <StockForm type="UPDATE"
                currentData={{
                  "corporation": stockInfo.corporation,
                  "name": stockInfo.name,
                  "symbol": stockInfo.symbol,
                  "shares": stockInfo.shares,
                  "lastUpdatedDate": new Date(reformatDate(stockInfo.lastUpdatedDate.toString(), '/', '-'))
                } as Stock}
                onChangeStockInfo={(newStockInfo) => setStockInfo(newStockInfo)}
                onHide={() => setUpdateStockFormPresent(false)} />
            </td>
          </tr>
      }

      { /* The form to delete the stock */
        deleteStockFormPresent && 
          <tr>
            <td>
              <DeleteStockForm stockSymbol={stockInfoData.symbol}
                onChangeStockList={onChangeStockList}
                onHide={() => setDeleteStockFormPresent(false)} />
            </td>
          </tr>
      }
    </>
  );
}

// the stock table  
function StockTable(): JSX.Element {
  const [addStockFormPresent, setAddStockFormPresent] = useState(false);
  const [stockList, setStockList] = useState<Stock[]>([]);

  // conditional rendering with useMediaQuery 
  const isMobileDevice = useMediaQuery("only screen and (max-width : 500px)");
  const isMediumDevice = useMediaQuery("only screen and (max-width : 1000px)");

  function fetchStockList(): void {
    getStocks()
      .then((response) => setStockList(response.data))
      .catch((error) => handleError(error));
  }
  useEffect(fetchStockList, []);

  return (
    <div className="stock-list">
      <div className="table-title">
        <h3>List of stocks ({stockList.length}):</h3>
      </div>

      <table>
        <thead>
          <tr>
            <th>Symbol</th>
            {!isMediumDevice && <th>Name</th>}
            <th>Price</th>
            <th>Shares Value</th>
            {!isMediumDevice && <th>Last updated</th>}
            {!isMobileDevice && (
              <>
                <th>More</th>
                <th>Update</th>
                <th>Delete</th>
              </>
            )}
          </tr>
        </thead>
        <tbody>
          {stockList.length === 0 && 
            <tr><td colSpan={isMobileDevice ? 3 : 8} style={{
              textAlign: "center",
              fontWeight: "bold",
              padding: "10px 0",
            }}>You don't have any stocks</td></tr>
          }
          {stockList.map(stock =>
            <StockTableRow
              key={stock.id}
              stockInfoData={stock}
              onChangeStockList={(newStockList: Stock[]) => setStockList(newStockList)} />
          )}
        </tbody>
      </table>
      <button style={{ display: "block", margin: "20px auto", }}
        onClick={() => setAddStockFormPresent(!addStockFormPresent)}>
        <FontAwesomeIcon icon={faPlusCircle} style={{ display: "block", margin: "0 auto 5px auto" }} />Add Stock
      </button>
      
      {addStockFormPresent && <StockForm 
        type='ADD'
        onChangeStockList={(newStockList) => setStockList(newStockList)}
        onHide={() => setAddStockFormPresent(false)} />}
    </div>
  );
}


interface InvestmentPageProps {
  navbarWidth: number, 
  titleHeight: number
}

/**
 * The investment page 
 * @param {InvestmentPageProps}
 * @returns {JSX.Element}
 */
export default function InvestmentPage({ navbarWidth, titleHeight }: InvestmentPageProps): JSX.Element {
  const [porfolioValueList, setPortfolioValueList] = useState<Record<string, number>>({});
  const [portfolioValue, setPortfolioValue] = useState(0);
  const [updatedDate, setUpdatedDate] = useState("");

  /**
   * Fetch the list of stocks and porfolio values 
   * @returns {void}
   */
  function fetchPortfolioValuelist(): void {
    getPorfolioValues()
      .then((response) => {
        // set the data for the list of porfolio value 
        setPortfolioValueList(response.data);
        setPortfolioValue(Object.values(response.data).at(-1) as number);
        setUpdatedDate(Object.keys(response.data).at(-1) as string);
      })
      .catch((error) => handleError(error));
  }
  useEffect(fetchPortfolioValuelist, []);

  return (
    <div className="investment-page" style={{
      marginTop: `${titleHeight}px`,
      marginLeft: `${navbarWidth}px`,
      width: `calc(100% - ${navbarWidth}px)`,
    }}>
      <h2 style={{ width: "85%", margin: "0 auto" }}>Your stock porfolio</h2>
      <div className="portfolio-value-info">
        <div style={{ margin: "0 auto", padding: "5px 20px", borderRadius: "10px" }}>
          <p>
            Porfolio's value: <strong>${portfolioValue}</strong> <span>(updated at {updatedDate})</span>
          </p>
        </div>
        <div className="portfolio-chart">
          <TimeseriesChart
            intervalType="month" contentType="stock price" stockSymbol="porfolio"
            timeSeriesObject={porfolioValueList} />
        </div>
      </div>
      <StockTable />
    </div>
  );
}

