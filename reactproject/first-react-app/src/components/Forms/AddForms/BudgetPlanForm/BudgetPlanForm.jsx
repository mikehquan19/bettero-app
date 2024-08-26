import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { faX } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { expenseappClient } from "../../../../provider/api";
import "./BudgetPlanForm.css";

// the form to add or update the budget plan 
const BudgetPlanForm = ({
    type = "ADD",
    intervalType = null, 
    existingData = null, 
    handleHideForm = null,
    handleChangePlanList = null}) => {

    // necessary components of useForm() hook 
    const { register, reset, handleSubmit, formState: { errors } } = useForm();
    const [percentageNotAddUp, setPrecentageNotAddUp] = useState(false); 

    // set the initial value for the form 
    useEffect(() => {
        if (existingData !== null) {
            reset({...existingData});
        } else {
            reset((previousData) => ({
                ...previousData, 
                interval_type: intervalType, 
            }));
        }
    }, []);

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

    // handle the submission of the form 
    const callAPI = (argType, argIntervalType, argData) => {
        if (argType === "ADD") {
            return expenseappClient.post("/budget/", argData); 
        } else if (argType === "UPDATE") {
            return expenseappClient.put(`budget/${argIntervalType}/`, argData); 
        } else {
            return; 
        }
    }

    const handleBudgetFormSubsmit = (data, e) => {
        e.preventDefault(); 
        console.log(data);
        // validate if the all categories add up 100%
        if (+data.grocery + +data.dining + +data.shopping + +data.bills + +data.gas + +data.others != 100) {
            setPrecentageNotAddUp(true);
        }
        else {
            setPrecentageNotAddUp(false); 
            // call the correct API method based on the form's type
            const apiIntervalType = existingData === null ? null : existingData.interval_type;
            callAPI(type, apiIntervalType, data)
                .then((response) => {
                    // log the data in for checking 
                    console.log("add or update budget plan request status: " + response.status);
                    handleChangePlanList(response.data);
                })
                .catch((error) => handleError(error));
            // hide the form 
            handleHideForm();
        }
    }

    // validation on the form 
    const registerOptions = {
        recurring_income: {
            required: "*Income is required", 
            min: {
                value: 0.01, 
                message: "*Income must be greater than 0", 
            }, 
            maxLength: {
                value: 12, 
                message: "*Income only has at most 12 digits", 
            }
        }, 
        portion_for_expense: {
            required: "*Total budget is required", 
            min: {
                value: 0, 
                message: "*Percentage must be greater than or equal 0", 
            },
            max: {
                value: 100, 
                message: "*Percentage must be less than or equal 100", 
            }
        }, 
        grocery: {
            required: "*Grocery budget is required", 
            min: {
                value: 0, 
                message: "*Percentage must be greater than or equal 0", 
            },
            max: {
                value: 100, 
                message: "*Percentage must be less than or equal 100", 
            }
        }, 
        dining: {
            required: "*Dining budget is required", 
            min: {
                value: 0, 
                message: "*Percentage must be greater than or equal 0", 
            },
            max: {
                value: 100, 
                message: "*Percentage must be less than or equal 100", 
            }
        }, 
        shopping: {
            required: "*Shopping budget is required", 
            min: {
                value: 0, 
                message: "*Percentage must be greater than or equal 0", 
            },
            max: {
                value: 100, 
                message: "*Percentage must be less than or equal 100", 
            }
        }, 
        bills: {
            required: "*Bills budget is required", 
            min: {
                value: 0, 
                message: "*Percentage must be greater than or equal 0", 
            },
            max: {
                value: 100, 
                message: "*Percentage must be less than or equal 100", 
            }
        }, 
        gas: {
            required: "*Gas budget is required", 
            min: {
                value: 0, 
                message: "*Percentage must be greater than or equal 0", 
            },
            max: {
                value: 100, 
                message: "*Percentage must be less than or equal 100", 
            }
        }, 
        others: {
            required: "*Others budget is required", 
            min: {
                value: 0, 
                message: "*Percentage must be greater than or equal 0", 
            },
            max: {
                value: 100, 
                message: "*Percentage must be less than or equal 100", 
            }
        }
    }

    return (
        <>
            <div className="budget-plan-form-wrapper">
                <h3 style={{
                        marginTop: "0",
                        textAlign: "center",
                    }}>{type} BUDGET PLAN</h3>
                <form id="budget-plan-form" onSubmit={handleSubmit(handleBudgetFormSubsmit)}>
                    <div className="form-field">
                        <label htmlFor="estimated-income">Estimated income ($):</label>
                        <small className="error-message">
                            {errors?.recurring_income && errors.recurring_income.message}
                        </small>
                        <input type="number" id="estimated-income" placeholder="Estimated income"
                            name="recurring_income" {...register("recurring_income", registerOptions.recurring_income)} />
                    </div>
                    <div className="form-field">
                        <label htmlFor="total-budget">Total expense portion (%):</label>
                        <small className="error-message">
                            {errors?.portion_for_expense && errors.portion_for_expense.message}
                        </small>
                        <input type="number" id="total-budget" placeholder="Total budget"
                            name="portion_for_expense" {...register("portion_for_expense", registerOptions.portion_for_expense)} />
                    </div>

                    {/* some instructions */}
                    <h4 style={{
                        textAlign: "center",
                    }}>Enter the portion of total budget for each category</h4>
                    <p className="error-message" style={{
                        textAlign: "center",
                    }}>{percentageNotAddUp && "*All categories don't add up to 100%"}</p>

                    <div className="form-field">
                        <label htmlFor="grocery">Grocery (%):</label>
                        <small className="error-message">
                            {errors?.grocery && errors.grocery.message}
                        </small>
                        <input type="number" id="grocery" placeholder="Grocery"
                            name="grocery" {...register("grocery", registerOptions.grocery)} />
                    </div>
                    <div className="form-field">
                        <label htmlFor="dining">Dining (%):</label>
                        <small className="error-message">
                            {errors?.dining && errors.dining.message}
                        </small>
                        <input type="number" id="dining" placeholder="Dining"
                            name="dining" {...register("dining", registerOptions.dining)} />
                    </div>
                    <div className="form-field">
                        <label htmlFor="shopping">Shopping (%):</label>
                        <small className="error-message">
                            {errors?.shopping && errors.shopping.message}
                        </small>
                        <input type="number" id="shopping" placeholder="Shopping"
                            name="shopping" {...register("shopping", registerOptions.shopping)} />
                    </div>
                    <div className="form-field">
                        <label htmlFor="bills">Bills (%):</label>
                        <small className="error-message">
                            {errors?.bills && errors.bills.message}
                        </small>
                        <input type="number" id="bills" placeholder="Bills"
                            name="bills" {...register("bills", registerOptions.bills)} />
                    </div>
                    <div className="form-field">
                        <label htmlFor="gas">Gas (%):</label>
                        <small className="error-message">
                            {errors?.gas && errors.gas.message}
                        </small>
                        <input type="number" id="gas" placeholder="Gas"
                            name="gas" {...register("gas", registerOptions.gas)} />
                    </div>
                    <div className="form-field">
                        <label htmlFor="others">Others (%):</label>
                        <small className="error-message">
                            {errors?.others && errors.others.message}
                        </small>
                        <input type="number" id="others" placeholder="Others"
                            name="others" {...register("others", registerOptions.others)} />
                    </div>
                </form>
                {/* The buttons for form */}
                <div className="plan-form-buttons">
                    <button form="budget-plan-form" type="submit">Submit</button>
                    <button type="button" onClick={handleHideForm}>Close</button>
                </div>
                <button className="x-button" onClick={handleHideForm}><FontAwesomeIcon icon={faX}/></button>
            </div>
            <div className="overlay"></div>
        </>
    );
}

export default BudgetPlanForm; 