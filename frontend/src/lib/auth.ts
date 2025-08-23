// lib/auth.ts
import { type UnaryInterceptor, type StreamInterceptor } from "grpc-web";

// JWT Token management
export const getJWTToken = (): string | null => {
  try {
    const token = localStorage.getItem('sortedchat.jwt');
    console.debug('Retrieved JWT token from localStorage:', token ? 'Token found' : 'No token');
    return token;
  } catch (error) {
    console.debug('Error retrieving JWT token from localStorage:', error);
    return null;
  }
};

// Unary interceptor to add JWT token to all requests
export const jwtUnaryInterceptor: UnaryInterceptor<any, any> = {
  intercept: (request, invoker) => {
    const metadata = request.getMetadata();
    const token = getJWTToken();
    
    if (token) {
      metadata["authorization"] = `Bearer ${token}`;
      console.debug('Added JWT token to request metadata for method:', request.getMethodDescriptor().getName());
    } else {
      console.debug('No JWT token available for request:', request.getMethodDescriptor().getName());
    }
    
    return invoker(request);
  },
};

// Stream interceptor to add JWT token to all streaming requests
export const jwtStreamInterceptor: StreamInterceptor<any, any> = {
  intercept: (request, invoker) => {
    const metadata = request.getMetadata();
    const token = getJWTToken();
    
    if (token) {
      metadata["authorization"] = `Bearer ${token}`;
      console.debug('Added JWT token to streaming request metadata for method:', request.getMethodDescriptor().getName());
    } else {
      console.debug('No JWT token available for streaming request:', request.getMethodDescriptor().getName());
    }
    
    return invoker(request);
  },
};

// Helper function to create gRPC client options with JWT authentication for both unary and streaming calls
export const createAuthenticatedClientOptions = () => ({
  unaryInterceptors: [jwtUnaryInterceptor],
  streamInterceptors: [jwtStreamInterceptor],
});
