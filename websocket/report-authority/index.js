const AWS = require("aws-sdk");
const ddb = new AWS.DynamoDB.DocumentClient();

exports.handler = async function (event, context) {
  let connection;
  const rtoOfficeName = event.queryStringParameters ? event.queryStringParameters.office_name : null;
  console.log(rtoOfficeName);

  if (!rtoOfficeName) {
    return {
      statusCode: 400,
      message: "Missing office_name parameter in the query string.",
    };
  }

  try {
    const getParams = {
      TableName: process.env.RTO_OFFICE_TABLE_ARN,
      Key: {
        office_name: rtoOfficeName,
      },
    };

    const getResult = await ddb.get(getParams).promise();
    connection = getResult.Item;
  } catch (err) {
    return {
      statusCode: 500,
      message: "Internal Server Error: Error in fetching from DB.",
    };
  }

  if (!connection) {
    return {
      statusCode: 404,
      message: `No connection found for office_name: ${rtoOfficeName}`,
    };
  }

  const callbackAPI = new AWS.ApiGatewayManagementApi({
    apiVersion: "2018-11-29",
    endpoint: event.requestContext.domainName + "/" + event.requestContext.stage,
  });

  const data = event.body.Data;
  console.log(event.body);

  try {
    if (connection.connectionId !== event.requestContext.connectionId) {
      await callbackAPI
        .postToConnection({ ConnectionId: connection.connectionId, Data: JSON.stringify(data) })
        .promise();
    }
  } catch (e) {
    console.log(e);
    return { statusCode: 500 };
  }

  return { statusCode: 200 };
};