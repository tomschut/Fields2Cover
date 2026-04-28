import SwaggerUI from 'swagger-ui-react'
import 'swagger-ui-react/swagger-ui.css'

export default function ApiDocsPage() {
  return (
    <div className="swagger-page">
      <SwaggerUI
        url="/openapi.yaml"
        deepLinking={true}
        tryItOutEnabled={true}
      />
    </div>
  )
}
