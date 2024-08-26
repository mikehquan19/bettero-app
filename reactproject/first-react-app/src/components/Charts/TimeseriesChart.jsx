import {
    Chart as ChartJS,
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    Title,
    Tooltip,
    Legend,
    TimeScale,
} from 'chart.js';
import { Line } from 'react-chartjs-2';
import ChartDataLabels from 'chartjs-plugin-datalabels';
import 'chartjs-adapter-date-fns';

// register the tool to be able to to use it in the hand-shakeable way 
ChartJS.register(CategoryScale, LinearScale, PointElement,
    LineElement, TimeScale, Title, Tooltip, Legend, ChartDataLabels);

// reformat the date for value of the date field in the form 
const reformatDate = (argDate) => {
    const dateArray = argDate.split("/");
    const month = dateArray[0];
    const day = dateArray[1];
    const year = dateArray[2];
    return (month + "-" + day + "-" + year);
}

// the chart showing that 
const TimeseriesChart = ({contentType = "expense", stockSymbol = null, intervalType = "month", timeseriesObject = null}) => {

    if (timeseriesObject !== null) {
        // labels and data of the timeseries chart 
        var timeseriesLabels = [];
        var timeseriesData = [];
        for (let i = 0; i < Object.keys(timeseriesObject).length; i++) {
            const thisDate = Object.keys(timeseriesObject)[i]; 
            timeseriesLabels.push(new Date(reformatDate(thisDate))); 
            timeseriesData.push(timeseriesObject[thisDate]);
        }
    }

    let chartTitle = "The chart of the user's expense this " + intervalType;
    if (contentType === "stock price") {
        `The timeseries chart of ${stockSymbol}'s price`
        chartTitle = [`The chart of ${stockSymbol}'s price`, "(starting last month)"];
    }
    
    // data of the chart 
    const chartData = {
        labels: timeseriesLabels,
        datasets: [
            {
                label: "expense on the given date",
                data: timeseriesData,
                borderColor: "rgb(0, 0, 0, 0.5)",
            }
        ],
    };

    // options of the chart
    const chartOptions = {
        // this is the gain control of the width and height of the canvas
        maintainAspectRatio: false,
        responsive: true,
        scales: {
            x: { // so that the horizontal axis will display unit of time
                type: 'time',
                time: {
                    unit: 'day',
                    unitStepSize: 1,
                    displayFormats: {
                        'day': 'MMM dd',
                    },
                },
                grid: {
                    // change the color of the vertical grid of white
                    color: "skyblue",
                },
            },
        },
        plugins: {
            // set the title for the chart 
            title: {
                display: true,
                color: 'black',
                text: chartTitle,
                font: {
                    size: 20,
                }
            }
        },
    };

    return <Line data={chartData} options={chartOptions} />
}

export default TimeseriesChart; 