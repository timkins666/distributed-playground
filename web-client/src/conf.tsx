export const gatewayHost = "http://localhost:8080"

export const gatewayUrl = (service: string, path: string) => {
    return `${gatewayHost}/${service}/${path}`
}