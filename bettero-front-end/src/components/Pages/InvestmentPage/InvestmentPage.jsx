import { useState, useEffect } from "react";
import { useMediaQuery } from "@uidotdev/usehooks";
import TimeseriesChart from "../../Charts/TimeseriesChart";
import { faPlusCircle } from '@fortawesome/free-solid-svg-icons';
import { expenseappClient } from "../../../provider/api";
import handleError from "../../../provider/handleError";
import './InvestmentPage.css';
import StockForm from "../../Forms/AddForms/StockForm";
import DeleteStockForm from "../../Forms/DeleteForms/DeleteStockForm";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

// each row in the table of stock
const StockTableRow = ({ stockInfoData, onChangeStockList }) => {
  const [stockInfo, setStockInfo] = useState({ ...stockInfoData });
  const [detailRowPresent, setDetailRowPresent] = useState(false);

  const [stockPriceObject, setStockPriceObject] = useState({});
  const [loadStockPrice, setLoadStockPrice] = useState(false);

  const [updateStockFormPresent, setUpdateStockFormPresent] = useState(false);
  const [deleteStockFormPresent, setDeleteStockFormPresent] = useState(false);

  // conditional rendering with useMediaQuery 
  const isMobileDevice = useMediaQuery("only screen and (max-width : 500px)");
  const isMediumDevice = useMediaQuery("only screen and (max-width : 1000px)");

  // reformat the date for value of the date field in the form 
  const reformatDate = (argDate) => {
    if (argDate !== null) {
      const dateArray = argDate.split("/");
      const month = dateArray[0];
      const day = dateArray[1];
      const year = dateArray[2];
      return (year + "-" + month + "-" + day);
    }
    return null;
  }

  const fetchPriceDetail = () => {
    // if the stock price hasn't been loaded yet, load it
    if (!loadStockPrice) {
      setLoadStockPrice(true);
      expenseappClient.get(`/stocks/${stockInfo.symbol}`)
        .then((response) => {
          setStockPriceObject(response.data.price_list);
        })
        .catch((error) => handleError(error));
    }
  }

  // compute the value and value change of shares of stock
  const sharesValue = (stockInfo.current_close * stockInfo.shares).toFixed(2);
  const sharesValueChange = (stockInfo.change * stockInfo.shares).toFixed(2);

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
        <td><span style={{ display: "block" }}>${stockInfo.current_close}</span> ({stockInfo.change > 0 && (<span>+</span>)
        } {stockInfo.change})</td>
        <td><span style={{ display: "block" }}>${sharesValue}</span> ({sharesValueChange > 0 && (<span>+</span>)
        } {sharesValueChange})</td>

        {!isMediumDevice && <td>{stockInfo.last_updated_date}</td>}
        {!isMobileDevice && (
          <>
            {/* the button to see details, update, or delete the stock */}
            <td><button onClick={() => {
              fetchPriceDetail(); // load the stock price first, if not yet
              setDetailRowPresent(!detailRowPresent);
            }}>{detailRowPresent ? (<span>Back</span>) : (<span>More</span>)}</button></td>
            <td><button onClick={() => setUpdateStockFormPresent(true)}>Update</button></td>
            <td><button onClick={() => setDeleteStockFormPresent(true)}>Delete</button></td>
          </>
        )}
      </tr>

      {detailRowPresent && (
        <tr className="detail-row"><td colSpan={8}>
          <div className="detail-row-wrapper">
            <h3 style={{ textAlign: "center", margin: "20px auto" }}>The stock's detail last updated on {stockInfo.last_updated_date} (updated everyday)</h3>
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
                <TimeseriesChart
                  timeseriesObject={stockPriceObject}
                  contentType="stock price"
                  stockSymbol={stockInfo.symbol} />
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

      {/* The form to update the stock's info */}
      {updateStockFormPresent && (<StockForm
        type="UPDATE"
        currentFormData={{
          "corporation": stockInfo.corporation,
          "name": stockInfo.name,
          "symbol": stockInfo.symbol,
          "shares": stockInfo.shares,
          "last_updated_date": reformatDate(stockInfo.last_updated_date),
        }}
        onChangeStockInfo={(newStockInfo) => setStockInfo(newStockInfo)}
        onHideForm={() => setUpdateStockFormPresent(false)} />)}

      {/* The form to delete the stock */}
      {deleteStockFormPresent && (<DeleteStockForm
        stockSymbol={stockInfoData.symbol}
        onChangeStockList={onChangeStockList}
        onHideForm={() => setDeleteStockFormPresent(false)} />)}
    </>
  );
}

// the stock table  
const StockTable = () => {
  const [addStockFormPresent, setAddStockFormPresent] = useState(false);
  const [stockList, setStockList] = useState([]);

  // conditional rendering with useMediaQuery 
  const isMobileDevice = useMediaQuery("only screen and (max-width : 500px)");
  const isMediumDevice = useMediaQuery("only screen and (max-width : 1000px)");

  const fetchStockList = () => {
    expenseappClient.get("/stocks")
      .then((response) => {
        setStockList(response.data);
      })
      .catch((error) => handleError(error));
  }
  useEffect(fetchStockList, []);

  return (
    <div className="stock-list">
      {/* the title of the table  */}
      <div className="table-title">
        <h3>List of stocks ({stockList.length}):</h3>
      </div>

      {/* the table presenting the list of stocks */}
      <table>
        {/* caption of the table */}
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
          {/* if the list doesn't have any stocks, present */
            stockList.length === 0 && (
              <tr><td colSpan={isMobileDevice ? 3 : 8} style={{
                textAlign: "center",
                fontWeight: "bold",
                padding: "10px 0",
              }}>You don't have any stocks</td></tr>
            )}
          {stockList.map(stock =>
            <StockTableRow
              key={stock.id}
              stockInfoData={stock}
              onChangeStockList={(newStockList) => setStockList(newStockList)} />
          )}
        </tbody>
      </table>
      <button style={{ display: "block", margin: "20px auto", }}
        onClick={() => setAddStockFormPresent(!addStockFormPresent)}>
        <FontAwesomeIcon icon={faPlusCircle} style={{ display: "block", margin: "0 auto 5px auto" }} />Add Stock
      </button>
      
      {addStockFormPresent && <StockForm
        onChangeStockList={(newStockList) => setStockList(newStockList)}
        onHideForm={() => setAddStockFormPresent(false)} />}
    </div>
  );
}

// the investment page 
const InvestmentPage = ({ navbarWidth, titleHeight }) => {
  const [porfolioValueList, setPortfolioValueList] = useState([]);
  const [portfolioValue, setPortfolioValue] = useState(0);
  const [updatedDate, setUpdatedDate] = useState("");

  // fetch the list of stocks and porfolio values 
  const fetchPortfolioValuelist = () => {
    expenseappClient.get("/portfolio_value")
      .then((response) => {
        // set the data for the list of porfolio value 
        setPortfolioValueList(response.data);
        setPortfolioValue(Object.values(response.data).at(-1));
        setUpdatedDate(Object.keys(response.data).at(-1));
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
            timeseriesObject={porfolioValueList}
            contentType="stock price"
            stockSymbol={"porfolio"}
          />
        </div>
      </div>
      <StockTable />
    </div>
  );
}

export default InvestmentPage;

