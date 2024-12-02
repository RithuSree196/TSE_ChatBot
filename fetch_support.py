from datetime import datetime, timezone, timedelta
from azure.cosmos import CosmosClient, exceptions

# Azure Cosmos DB details
url = "https://cdbazewp-admiral-core.documents.azure.com:443/"
key = "ZZ5NJmqjAobGMI8w1MHbO4LN4oigpQnM0xwn0q40B9uoHYFqlgRvM6sEhdGv2UxpnJ0nqI1yLRuw2jc5Jx0s4w=="
database_name = "Admiral"
container_name = "Support"

# Initialize Cosmos Client and container
client = CosmosClient(url, credential=key)
database = client.get_database_client(database_name)
container = database.get_container_client(container_name)

# Calculate the timestamp for 90 days ago
ninety_days_ago = (datetime.now(timezone.utc) - timedelta(days=90)).isoformat()

# Query to fetch ticket details created in the last 90 days
details_query = f"""
SELECT * FROM c 
WHERE c.Data.createdDate <= GetCurrentDateTime() 
AND c.Data.createdDate >= '{ninety_days_ago}'
"""

try:
    # Execute the query to fetch ticket details
    ticket_details = list(container.query_items(
        query=details_query,
        enable_cross_partition_query=True
    ))
    
    # Process and save the results
    if ticket_details:
        # Save the fetched tickets to a Markdown file
        output_file = "support_tickets_last_90_days.md"
        with open(output_file, "w", encoding="utf-8") as md_file:
            md_file.write("# Support Tickets (Last 90 Days)\n\n")
            for ticket in ticket_details:
                md_file.write(f"## Ticket ID: {ticket['id']}\n")
                md_file.write(f"- **Created Date:** {ticket['Data']['createdDate']}\n")
                md_file.write(f"- **Details:** {ticket['Data']}\n\n")
        
        print(f"Ticket details from the last 90 days saved to {output_file}")
    else:
        print("No tickets found created in the last 90 days.")

except exceptions.CosmosHttpResponseError as e:
    print(f"An error occurred: {e.message}")
