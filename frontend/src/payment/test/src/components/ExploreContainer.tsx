import React, { useState, useEffect } from 'react';
import './ExploreContainer.css';
import IAP, { Product, Purchase } from '../payment/PaymentPlugin';

interface ContainerProps {
  name: string;
}

const ExploreContainer: React.FC<ContainerProps> = ({ name }) => {
  const [status, setStatus] = useState<string>('Not initialized');
  const [products, setProducts] = useState<Product[]>([]);
  const [purchases, setPurchases] = useState<Purchase[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  useEffect(() => {
    initializePayment();
  }, []);

  const initializePayment = async () => {
    try {
      setStatus('Initializing...');
      const result = await IAP.initialize();
      if (result.success) {
        setStatus('Initialized');
        await loadProducts();
      } else {
        setStatus('Failed to initialize');
      }
    } catch (error) {
      setStatus('Error initializing: ' + error);
    }
  };

  const loadProducts = async () => {
    try {
      const result = await IAP.getProducts({ productIds: ['prod_1', 'prod_2'] });
      setProducts(result.products);
    } catch (error) {
      setStatus('Error loading products: ' + error);
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

      {products.length > 0 && (
        <div style={{ marginTop: '20px' }}>
          <h4>Available Products:</h4>
          {products.map((product) => (
            <div key={product.productId} style={{ 
              border: '1px solid #ccc', 
              padding: '10px', 
              margin: '10px 0',
              borderRadius: '5px'
            }}>
              <h5>{product.title}</h5>
              <p>{product.description}</p>
              <p>Price: ${product.priceAmount} {product.currency}</p>
              <button 
                onClick={() => handlePurchase(product.productId)}
                disabled={isLoading}
                style={{
                  padding: '10px 20px',
                  backgroundColor: '#007bff',
                  color: 'white',
                  border: 'none',
                  borderRadius: '5px',
                  cursor: isLoading ? 'not-allowed' : 'pointer'
                }}
              >
                {isLoading ? 'Processing...' : 'Buy Now'}
              </button>
            </div>
          ))}
        </div>
      )}

      {purchases.length > 0 && (
        <div style={{ marginTop: '20px' }}>
          <h4>Your Purchases:</h4>
          {purchases.map((purchase) => (
            <div key={purchase.transactionId} style={{ 
              border: '1px solid #28a745', 
              padding: '10px', 
              margin: '10px 0',
              borderRadius: '5px',
              backgroundColor: '#f8f9fa'
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
