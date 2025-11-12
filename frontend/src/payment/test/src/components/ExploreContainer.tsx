import React, { useState } from 'react';
import './ExploreContainer.css';
import IAP, { Purchase, Product } from '../payment/PaymentPlugin';

interface ContainerProps {
  name: string;
}

const ExploreContainer: React.FC<ContainerProps> = ({ name }) => {
  const [status, setStatus] = useState<string>('Not initialized');
  const [purchases, setPurchases] = useState<Purchase[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const initializePayment = async () => {
    try {
      setStatus('Initializing...');
      setIsLoading(true);
      const result = await IAP.initialize();
      if (result.success) {
        setStatus('Ready for payment');
      } else {
        setStatus('Failed to initialize');
      }
    } catch (error) {
      setStatus('Error initializing: ' + error);
    } finally {
      setIsLoading(false);
    }
  };

  const showProducts = async () => {
    try {
      setStatus('Loading products...');
      setIsLoading(true);
      // Get products with some example product IDs
      const result = await IAP.getProducts({ 
        productIds: ['exampleproduct1'] 
      });
      setProducts(result.products);
      setStatus(`Found ${result.products.length} products`);
    } catch (error) {
      setStatus('Error loading products: ' + error);
    } finally {
      setIsLoading(false);
    }
  };

  const handlePurchase = async (productId: string) => {
    setIsLoading(true);
    try {
      setStatus('Processing purchase...');
      const result = await IAP.purchaseProduct({ productId });
      
      if (result.purchase) {
        setStatus(`Purchase successful! Transaction: ${result.purchase.transactionId}`);
        // Refresh purchases
        const purchasesResult = await IAP.getPurchases();
        setPurchases(purchasesResult.purchases);
      } else {
        setStatus('Purchase failed: ' + result.message);
      }
    } catch (error) {
      setStatus('Purchase error: ' + error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div id="container" style={{ padding: '20px' }}>
      <strong>{name}</strong>
      
      <div style={{ marginTop: '20px' }}>
        <h3>Payment Status: {status}</h3>
      </div>

      {/* Three main buttons */}
      <div style={{ marginTop: '20px', display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
        <button 
          onClick={initializePayment}
          disabled={isLoading}
          style={{
            padding: '15px 30px',
            backgroundColor: isLoading ? '#6c757d' : '#28a745',
            color: 'white',
            border: 'none',
            borderRadius: '8px',
            cursor: isLoading ? 'not-allowed' : 'pointer',
            fontSize: '16px',
            fontWeight: 'bold'
          }}
        >
          {isLoading ? 'Initializing...' : 'Initialize'}
        </button>

        <button 
          onClick={showProducts}
          disabled={isLoading}
          style={{
            padding: '15px 30px',
            backgroundColor: isLoading ? '#6c757d' : '#17a2b8',
            color: 'white',
            border: 'none',
            borderRadius: '8px',
            cursor: isLoading ? 'not-allowed' : 'pointer',
            fontSize: '16px',
            fontWeight: 'bold'
          }}
        >
          {isLoading ? 'Loading...' : 'Show Products'}
        </button>
      </div>

      {/* Products list */}
      {products.length > 0 && (
        <div style={{ marginTop: '20px' }}>
          <h4>Available Products:</h4>
          {products.map((product) => (
            <div key={product.productId} style={{ 
              border: '1px solid #ccc', 
              padding: '20px', 
              margin: '10px 0',
              borderRadius: '10px',
              backgroundColor: '#f8f9fa'
            }}>
              <h5>{product.title}</h5>
              <p>{product.description}</p>
              <p style={{ fontSize: '18px', fontWeight: 'bold', color: '#28a745' }}>
                Price: ${product.priceAmount} {product.currency}
              </p>
              <button 
                onClick={() => handlePurchase(product.productId)}
                disabled={isLoading}
                style={{
                  padding: '15px 30px',
                  backgroundColor: isLoading ? '#6c757d' : '#007bff',
                  color: 'white',
                  border: 'none',
                  borderRadius: '8px',
                  cursor: isLoading ? 'not-allowed' : 'pointer',
                  fontSize: '16px',
                  fontWeight: 'bold'
                }}
              >
                {isLoading ? 'Processing...' : 'Pay Now'}
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Purchases list */}
      {purchases.length > 0 && (
        <div style={{ marginTop: '20px' }}>
          <h4>Your Purchases:</h4>
          {purchases.map((purchase) => (
            <div key={purchase.transactionId} style={{ 
              border: '1px solid #28a745', 
              padding: '10px', 
              margin: '10px 0',
              borderRadius: '5px',
              backgroundColor: '#d4edda'
            }}>
              <p>Product: {purchase.productId}</p>
              <p>Transaction: {purchase.transactionId}</p>
              <p>Status: {purchase.state}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default ExploreContainer;
