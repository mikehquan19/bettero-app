import { useState, useEffect } from 'react';
import { useMediaQuery } from '@uidotdev/usehooks';
import CompositionChart from '../../Charts/CompositionChart';
import BudgetPlanForm from '../../Forms/AddForms/BudgetPlanForm/BudgetPlanForm';
import DeletePlanForm from '../../Forms/DeleteForms/DeletePlanForm/DeletePlanForm';
import BillForm from '../../Forms/AddForms/BillForm/BillForm';
import DeleteBillForm from '../../Forms/DeleteForms/DeleteBillForm/DeleteBillForm';
import { expenseappClient } from '../../../provider/api';
import { faPlusCircle, faCaretDown, faX } from '@fortawesome/free-solid-svg-icons';
import './BudgetPage.css';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

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


// function to log the error in the console depending on the type of error 
const handleError = (error) => {
    if (error.response) {
        // The request was made and the server responded with a status code
        // that falls out of the range of 2xx
        console.log(error.response.data);
        console.log(error.response.status);
        console.log(error.response.headers);
    } else if (error.request) {
        // The request was made but no response was received
        console.log(error.request);
    } else {
        // Something happened in setting up the request that triggered an Error
        console.log('Error', error.message);
    }
    console.log(error.config);
}


const BillTableRow = ({billInfo, onChangeBillList}) => {
    const [billFormPresent, setBillFormPresent] = useState(false); 
    const [deleteBillFormPresent, setDeleteBillFormPresent] = useState(false);
    const [detailRowPresent, setDetailRowPresent] = useState(false); 
    const isMobileDevice = useMediaQuery("only screen and (max-width : 500px)");
    const isMediumDevice = useMediaQuery("only screen and (max-width : 1000px)");

    return (
        <>
            <tr onClick={() => {
                if (isMobileDevice) {
                    setDetailRowPresent(!detailRowPresent); 
                }
            }}>
                <td>{billInfo.description}</td>
                {(!isMediumDevice || isMobileDevice) && <td>${billInfo.amount}</td>}
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
                        <p><strong>Pay account:</strong> {billInfo.pay_account_name}</p>
                        <p><strong>Due date:</strong> {billInfo.due_date}</p>
                        {/* the button to delete the bills */}
                        <div style={{display: "flex", flexDirection: "row", gap: "5px", justifyContent: "center"}}>
                            <button className="bill-button"
                                onClick={() => setDetailRowPresent(false)}>Back</button>
                            <button className="bill-button"
                                onClick={() => setDeleteBillFormPresent(true)}>Delete</button>
                            <button className="bill-button"
                                onClick={() => setBillFormPresent(true)}>Update</button>
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
                    pay_account: billInfo.pay_account,
                    due_date: reformatDate(billInfo.due_date),
                }}
                onChangeBillList={onChangeBillList}
                onHideForm={() => setBillFormPresent(false)}/>)}
            {deleteBillFormPresent && (<DeleteBillForm 
                billId={billInfo.id}
                onChangeBillList={onChangeBillList}
                onHideForm={() => setDeleteBillFormPresent(false)} />)}
        </>
    ); 
}


// the card showing the information of the bills 
const BillTable = () => {
    // the list of bills to be displayed
    // list of titles 
    const [billList, setBillList] = useState([]); 
    const [billFormPresent, setBillFormPresent] = useState(false); 
    const isMobileDevice = useMediaQuery("only screen and (max-width : 500px)");
    const isMediumDevice = useMediaQuery("only screen and (max-width : 1000px)");
    
    // function to fetch the user's list of bills
    const fetchBillList = () => {
        expenseappClient.get("/bills/")
            .then((response) => {
                // for checking, log the list of bills in 
                console.log("get bills request status: " + response.status);
                setBillList(response.data);
            })
            .catch((error) => handleError(error));
    }
    // fetch data only once 
    useEffect(fetchBillList, []);

    return (
        <div className="bills-list">
            {/* The title of the table of bill's list */}
            <div className="table-title">
                <h3 style={{
                    display: "inline-block",
                    margin: "0",
                }}>Bills this month ({billList.length})</h3>
            </div>
            {/* The table of list of bills */}
            <table>
                <thead>
                    <tr>
                        <th>Description</th>
                        {(!isMediumDevice || isMobileDevice) && <th>Amount</th>}
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
                    {   // rendering the messages when there are no bills 
                        billList.length === 0 && (
                            <tr><td colSpan={isMobileDevice ? 3 : 6}>You haven't added any bills</td></tr>
                        )
                    }
                    {
                        // if the length of the list of bills is 0, 
                        // it won't render anything 
                        billList.map(bill => 
                            <BillTableRow 
                                key={bill.id}
                                billInfo={bill}
                                onChangeBillList={(newBillList) => setBillList(newBillList)} />)
                    }
                </tbody>
            </table>
            {/* The button to add the new bills for this month */}
            <button className="add-bill-button" 
            onClick={() => setBillFormPresent(true)}>
                <FontAwesomeIcon icon={faPlusCircle} style={{display: "block", margin: "0 auto 5px auto"}}/>Add Bill
            </button>
            {/* The form to add bill */}
            {billFormPresent && (<BillForm 
                type="ADD"
                onChangeBillList={(newBillList) => setBillList(newBillList)}
                onHideForm={() => setBillFormPresent(false)}/>)}
        </div>
    ); 
}


// the progress bar 
const ProgressBar = ({ categoryExpenseData }) => {
    const [showCurrent, setShowCurrent] = useState(false);
    const [categoryInfo, setCategoryInfo] = useState({
        name: "",
        budget: 0,
        current: 0,
        percentage: 0,
    });

    useEffect(() => {
        setCategoryInfo({ ...categoryExpenseData });
    }, [categoryExpenseData]);

    return (
        <div className="budget-bar">
            <label htmlFor="progress-part">
                <strong>{categoryInfo.name}:</strong>
            </label>
            <div className="number-part">
                <span className="display-budget">
                    Budget: <strong>${categoryInfo.budget.toFixed(2)}</strong>
                </span>
                <span className="display-left">
                    <strong>${(categoryInfo.budget - categoryInfo.current).toFixed(2)}</strong> left
                </span>
            </div>
            {categoryInfo.current >= categoryInfo.budget && (
                <small style={{
                    color: "red",
                    display: "block",
                    marginBottom: "10px",
                }}>You have used up all of your budget of this category. Be more cautious!</small>
            )}
            <div className="progress-bar">
                <div id="progress-part" className="filled-area" style={{
                    width: `${categoryInfo.percentage}%`,
                }}></div>
            </div>
            <span className="percentage-part" style={{
                left: `${categoryInfo.percentage - 4}%`,
            }}
                onMouseOver={() => setShowCurrent(!showCurrent)}
                onMouseOut={() => setShowCurrent(!showCurrent)}>
                <strong>
                    {showCurrent ?
                        <span>${categoryInfo.current.toFixed(2)}</span> :
                        <span>{categoryInfo.percentage.toFixed(2)}%</span>
                    }
                </strong>
            </span>
        </div>
    );
}


// the panel showing the budget plan of the user 
const BudgetPlanPanel = ({ 
    intervalType, 
    estimatedIncome, 
    budgetPecentage, 
    budgetChartData, 
    actualChartData, 
    progressData, 
    handleChangePlanList = null}) => {

    // whether the form is present or not 
    const [addFormPresent, setAddFormPresent] = useState(false); 
    const [updateFormPresent, setUpdateFormPresent] = useState(false); 
    const [deleteFormPresent, setDeleteFormPresent] = useState(false);

    // whether one button is disabled because the other one is 
    const [updateButtonDisabled, setUpdateButtonDisabled] = useState(false); 
    const [deleteButtonDisabled, setDeleteButtonDisabled] = useState(false); 

    // everytime the user changes interval type, reset everything
    useEffect(() => {
        setAddFormPresent(false); 
        setUpdateFormPresent(false);
        setUpdateButtonDisabled(false);
        setDeleteButtonDisabled(false);
    }, [intervalType]);

    // prepare the data that will be passed to the update form 
    let existingData; 
    if (Object.keys(budgetChartData).length !== 0) {
        existingData = {
            interval_type: intervalType === "bi-week" ? "bi_week" : intervalType, 
            recurring_income: estimatedIncome, 
            portion_for_expense: budgetPecentage, 
            grocery: budgetChartData.Grocery,
            dining: budgetChartData.Dining,
            shopping: budgetChartData.Shopping,
            bills: budgetChartData.Bills,
            gas: budgetChartData.Gas,
            others: budgetChartData.Others,
        }
    }  
    
    return (
        <>
            <h2 className="budget-plan-header">The budget plan for the {intervalType}:</h2>
            <div className="budget-plan-panel">
                {Object.keys(budgetChartData).length === 0 ? (
                    <>
                        <h3 className='none-message'>You currently don't have a plan for the {intervalType}</h3>
                        <button 
                            className="add-plan-button"
                            onClick={() => setAddFormPresent(!addFormPresent)}>
                             <FontAwesomeIcon icon={faPlusCircle} style={{display: "block", margin: "0 auto 5px auto"}}/>Add Plan
                        </button>
                        {addFormPresent && <BudgetPlanForm 
                            type="ADD" 
                            intervalType={intervalType === "bi-week" ? "bi_week" : intervalType}
                            handleHideForm={() => setAddFormPresent(!addFormPresent)}
                            handleChangePlanList={handleChangePlanList} />}
                    </>
                ) : (
                    <>
                        <div className="income-and-budget">
                            <h3>Your estimated income this {intervalType}: ${estimatedIncome}</h3>
                            <h3>You put {budgetPecentage}% of your income on expense</h3>
                        </div>
                        <div className="composition-charts-part">
                            <div>
                                <CompositionChart chartType="budget" chartWidth="410"
                                    intervalType={intervalType} 
                                    compositionObject={budgetChartData} /> 
                            </div>
                            <div>
                                <CompositionChart chartType="actual" chartWidth="410"
                                    intervalType={intervalType} 
                                    compositionObject={actualChartData} />
                            </div>
                        </div>
                        <div className="progress-bars-part">
                            {Object.keys(progressData).map(category => 
                                <ProgressBar key={category} categoryExpenseData={{
                                    name: category,
                                    budget: progressData[category].budget, 
                                    current: progressData[category].current, 
                                    percentage: progressData[category].percentage,
                                }} />
                            )}
                        </div>
                        <div className="budget-button-list">
                            <button disabled={updateButtonDisabled} onClick={() => {
                                setUpdateFormPresent(!updateFormPresent);
                                setDeleteButtonDisabled(!deleteButtonDisabled);
                            }}>Update</button>
                            <button disabled={deleteButtonDisabled} onClick={() => {
                                setDeleteFormPresent(!deleteFormPresent);
                                setUpdateButtonDisabled(!updateButtonDisabled); 
                            }}>Delete</button>
                        </div>
                        {updateFormPresent && (<BudgetPlanForm 
                            type="UPDATE" 
                            existingData={existingData}
                            handleHideForm={() => {
                                setUpdateFormPresent(!updateFormPresent);
                                setDeleteButtonDisabled(!deleteButtonDisabled);
                            }} 
                            handleChangePlanList={handleChangePlanList} />)
                        }
                        {deleteFormPresent && (<DeletePlanForm 
                            intervalType={intervalType === "bi-week" ? "bi_week" : intervalType}
                            handleHideForm={() => {
                                setDeleteFormPresent(!deleteFormPresent);
                                setUpdateButtonDisabled(!updateButtonDisabled); 
                            }}
                            handleChangePlanList={handleChangePlanList} />)
                        }
                    </>
                )}
            </div>
        </>
    );
}

// the entire budget page 
const BudgetPage = ({ navbarWidth, titleHeight }) => {

    // fetch data from the api 
    const [budgetData, setBudgetData] = useState({
        // initial data just for the sake of initial rendering 
        month: { income: 0, expense_portion: 0, composition: { goal: {}, actual: {} }, progress: {} },
        bi_week: { income: 0, expense_portion: 0, composition: { goal: {}, actual: {} }, progress: {} },
        week: { income: 0, expense_portion: 0, composition: { goal: {}, actual: {} }, progress: {} },
    });
    const [overdueMessageList, setOverdueMessageList] = useState([]);
    const [overdueMessageListPresent, setOverdueMessageListPresent] = useState(false);
    const [caretIconStyle, setCaretIconStyle] = useState({
        marginLeft: "10px", 
        fontSize: "22px",
        transition: "transform 0.5s linear",
    });
    useEffect(() => {
        Promise.all([
            expenseappClient.get(`/budget/`), // the list of budgets of the user 
            expenseappClient.get(`/overdue_message/`) // the list of overdue bill message
        ])
            .then((response) => {
                console.log("get budget and overdue message request status: " + response.status)
                setBudgetData(response[0].data);
                setOverdueMessageList(response[1].data); 
            })
            .catch((error) => handleError(error)); 
    }, []);

    // state of the form 
    const [intervalType, setIntervalType] = useState("month");
    const [buttonColors, setButtonColors] = useState({
        month: "rgba(0, 192, 192, 1)",
        biWeek: "rgba(0, 192, 192, 0)",
        week: "rgba(0, 192, 192, 0)",
    });

    // update chart data 
    let budgetChartData = {};
    let actualChartData = {};
    let progressData = {};
    let estimatedIncome = 0; 
    let budgetPecentage = 0; 
    if (Object.keys(budgetData[intervalType]).length !== 0) {
        budgetChartData = budgetData[intervalType].composition.goal;
        actualChartData = budgetData[intervalType].composition.actual;
        progressData = budgetData[intervalType].progress;
        estimatedIncome = budgetData[intervalType].income; 
        budgetPecentage = budgetData[intervalType].expense_portion
    }
    // for checking 
    console.log("Budget plan for " + intervalType);
    console.log(budgetData[intervalType].progress);

    // type to be display
    const displayedType = intervalType === "bi_week" ? "bi-week" : intervalType;

    return (
        <div className="budget-page" style={{
            marginTop: `${titleHeight}px`,
            marginLeft: `${navbarWidth}px`,
            width: `calc(100% - ${navbarWidth}px)`,
        }}>
            <h2 style={{
                width: "80%",
                margin: "0 auto",
            }}>List of your monthly bills and budget plans</h2>

            {overdueMessageList.length !== 0 && (
                <div className="overdue-message-list">
                    <h3 onClick={() => {
                        if (!overdueMessageListPresent) {
                            setOverdueMessageListPresent(true);
                            setCaretIconStyle({
                                marginLeft: "10px", 
                                fontSize: "20px",
                                transform: "rotate(180deg)",
                                transition: "transform 0.5s linear",
                            });
                        } else {
                            setOverdueMessageListPresent(false);
                            setCaretIconStyle({
                                marginLeft: "10px", 
                                fontSize: "20px",
                                transition: "transform 0.5s linear",
                            });
                        }
                    }}>
                        **You have {overdueMessageList.length} overdue bills 
                        <FontAwesomeIcon icon={faCaretDown} style={caretIconStyle} />
                    </h3>
                    {overdueMessageListPresent && (
                        <div className="overdue-message-wrapper">
                            {overdueMessageList.map((message, index) => 
                                <div className="overdue-message" key={index}>
                                    <p><strong>{index + 1}/</strong> {message.bill_description}: ${message.bill_amount}, due on <strong>{message.bill_due_date}</strong></p>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            )}

            {/* The table of list of bills  */}
            <div className="bills-table">
                <BillTable />
            </div>

            {/* The panel consisting of buttons of interval type */}
            <div className="interval-button-wrapper">
                <div className="interval-button-list">
                    <button
                        onClick={() => {
                            setButtonColors({
                                month: "rgba(0, 192, 192, 1)",
                                biWeek: "rgba(0, 192, 192, 0)",
                                week: "rgba(0, 192, 192, 0)",
                            });
                            setIntervalType("month");
                        }}
                        style={{ backgroundColor: buttonColors.month, }}>Month</button>
                    <button
                        onClick={() => {
                            setButtonColors({
                                month: "rgba(0, 192, 192, 0)",
                                biWeek: "rgba(0, 192, 192, 1)",
                                week: "rgba(0, 192, 192, 0)",
                            });
                            setIntervalType("bi_week");
                        }}
                        style={{ backgroundColor: buttonColors.biWeek, }}>Bi Week</button>
                    <button
                        onClick={() => {
                            setButtonColors({
                                month: "rgba(0, 192, 192, 0)",
                                biWeek: "rgba(0, 192, 192, 0)",
                                week: "rgba(0, 192, 192, 1)",
                            });
                            setIntervalType("week");
                        }}
                        style={{ backgroundColor: buttonColors.week, }}>Week</button>
                </div>
            </div>

            {/* The panel showing the budget plan */}
            <BudgetPlanPanel
                intervalType={displayedType}
                estimatedIncome={estimatedIncome} budgetPecentage={budgetPecentage}
                budgetChartData={budgetChartData} actualChartData={actualChartData}
                progressData={progressData} handleChangePlanList={(newBudgetData) => setBudgetData(newBudgetData)} />
        </div>
    );
}

export default BudgetPage;